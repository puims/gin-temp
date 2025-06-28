package util

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/casbin/casbin/v2"
	gormAdapter "github.com/casbin/gorm-adapter/v3"
	"github.com/fsnotify/fsnotify"
	"gorm.io/gorm"
)

type CasbinPolicyLoader struct {
	enforcer *casbin.Enforcer
	path     string
	watcher  *fsnotify.Watcher
}

func (ld *CasbinPolicyLoader) loadPolicy() (err error) {
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
			log.Printf("Invalid policy line: %s", line)
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

	log.Println("Casbin policy reloaded from file")
	return nil
}

func (ld *CasbinPolicyLoader) watchPolicyFile() {
	if err := ld.watcher.Add(ld.path); err != nil {
		log.Printf("Failed to watch policy file: %v", err)
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
					log.Printf("Policy file modified, reloading...")
					if err := ld.loadPolicy(); err != nil {
						log.Printf("Failed to reload policy: %v", err)
					}
				})
			}
		case err, ok := <-ld.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Policy file watcher error: %v", err)
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

func newPolicyLoader(enforcer *casbin.Enforcer, path string) (*CasbinPolicyLoader, error) {

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	loader := &CasbinPolicyLoader{
		enforcer: enforcer,
		path:     path,
		watcher:  watcher,
	}

	if err := loader.loadPolicy(); err != nil {
		return nil, err
	}

	go loader.watchPolicyFile()

	return loader, nil
}

func setupCasbin(db *gorm.DB) (*casbin.Enforcer, *CasbinPolicyLoader, error) {
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

	loadPath := fmt.Sprintf("%s/.%s/.policy", home, Viper.GetString("app.name"))
	loader, err := newPolicyLoader(enforcer, loadPath)
	if err != nil {
		return nil, nil, err
	}

	return enforcer, loader, nil
}
