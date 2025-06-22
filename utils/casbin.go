package utils

import (
	"bufio"
	"fmt"
	"gin-temp/config"
	"os"
	"strings"
	"time"

	"github.com/casbin/casbin/v2"
	gormAdapter "github.com/casbin/gorm-adapter/v3"
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type CasbinPolicyLoader struct {
	enforcer *casbin.Enforcer
	path     string
	watcher  *fsnotify.Watcher
}

func (ld *CasbinPolicyLoader) LoadPolicy() (err error) {
	file, err := os.Open(ld.path)
	if err != nil {
		return err
	}
	defer file.Close()

	ld.enforcer.ClearPolicy()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if !ld.validatePolicyLine(line) {
			logrus.Warnf("Invalid policy line: %s", line)
			continue
		}

		parts := strings.Split(line, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}

		switch parts[0] {
		case "p":
			if len(parts) >= 4 {
				ld.enforcer.AddPolicy(parts[1:])
			}
		case "g":
			if len(parts) >= 3 {
				ld.enforcer.AddGroupingPolicy(parts[1:])
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if ld.enforcer.SavePolicy(); err != nil {
		return err
	}

	logrus.Info("Casbin policy reloaded from file")
	return nil
}

func (ld *CasbinPolicyLoader) watchPolicyFile() {
	if err := ld.watcher.Add(ld.path); err != nil {
		logrus.Errorf("Failed to watch policy file: %v", err)
		return
	}

	var timer *time.Timer
	for {
		select {
		case event, ok := <-ld.watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write ||
				event.Op&fsnotify.Create == fsnotify.Create {
				if timer != nil {
					timer.Stop()
				}

				timer = time.AfterFunc(500*time.Millisecond, func() {
					logrus.Info("Policy file modified, reloading...")
					if err := ld.LoadPolicy(); err != nil {
						logrus.Errorf("Failed to reload policy: %v", err)
					}
				})
			}
		case err, ok := <-ld.watcher.Errors:
			if !ok {
				return
			}
			logrus.Errorf("Policy file watcher error: %v", err)
		}
	}
}

func (*CasbinPolicyLoader) validatePolicyLine(line string) bool {
	parts := strings.Split(line, ",")
	if len(parts) == 0 {
		return false
	}

	switch parts[0] {
	case "p":
		return len(parts) >= 4
	case "g":
		return len(parts) >= 3
	default:
		return false
	}
}

func (ld *CasbinPolicyLoader) Close() error {
	return ld.watcher.Close()
}

func NewPolicyLoader(enforcer *casbin.Enforcer, path string) (*CasbinPolicyLoader, error) {

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	loader := &CasbinPolicyLoader{
		enforcer: enforcer,
		path:     path,
		watcher:  watcher,
	}

	if err := loader.LoadPolicy(); err != nil {
		return nil, err
	}

	go loader.watchPolicyFile()

	return loader, nil
}

func SetupCasbin(db *gorm.DB) (*casbin.Enforcer, *CasbinPolicyLoader, error) {
	adapter, err := gormAdapter.NewAdapterByDB(db)
	if err != nil {
		return nil, nil, err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, nil, err
	}

	enforcer, err := casbin.NewEnforcer("config/rbac.env", adapter)
	if err != nil {
		return nil, nil, err
	}

	loadPath := fmt.Sprintf("%s/.%s/.policy", home, config.Viper.GetString("app.name"))
	loader, err := NewPolicyLoader(enforcer, loadPath)
	if err != nil {
		return nil, nil, err
	}

	return enforcer, loader, nil
}
