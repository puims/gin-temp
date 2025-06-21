package models

import (
	"errors"
	"fmt"
	"gin-temp/config"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type MysqlDB struct {
	*gorm.DB
}

// NewMysqlDB 创建新的数据库连接实例
func NewMysqlDB(tables ...interface{}) (*MysqlDB, error) {
	user := config.Viper.GetString("mysql.user")
	pwd := config.Viper.GetString("mysql.password")
	host := config.Viper.GetString("mysql.host")
	port := config.Viper.GetInt("mysql.port")
	dbname := config.Viper.GetString("mysql.db")

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, pwd, host, port, dbname,
	)

	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       dsn,
		DefaultStringSize:         256,
		DisableDatetimePrecision:  true,
		DontSupportRenameIndex:    true,
		DontSupportRenameColumn:   true,
		SkipInitializeWithVersion: false,
	}), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            config.Viper.GetBool("mysql.prepareStmt"),
		Logger:                 logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}

	// 设置创建后的回调
	db.Callback().Create().After("gorm:create").Register("update_created_at", func(db *gorm.DB) {
		if db.Statement.Schema != nil {
			if field := db.Statement.Schema.LookUpField("CreatedAt"); field != nil {
				now := time.Now()
				db.Statement.SetColumn("CreatedAt", now, true)
			}
		}
	})

	// 设置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	maxLifeTime := config.Viper.GetInt("mysql.maxLifeTime")
	sqlDB.SetMaxOpenConns(config.Viper.GetInt("mysql.maxOpenConns"))
	sqlDB.SetMaxIdleConns(config.Viper.GetInt("mysql.maxIdleConns"))
	sqlDB.SetConnMaxLifetime(time.Duration(maxLifeTime) * time.Minute)

	mysqlDB := &MysqlDB{db}

	mysqlDB.Migrate(tables...)
	mysqlDB.CreateDefaultRoles()

	if err := mysqlDB.Ping(); err != nil {
		log.Fatalf("Database ping failed: %v", err)
	}

	return mysqlDB, nil
}

// Close 关闭数据库连接
func (db *MysqlDB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Migrate 执行数据库迁移
func (db *MysqlDB) Migrate(tables ...interface{}) error {
	mig := db.Migrator()
	for _, model := range tables {
		if !mig.HasTable(model) {
			if err := mig.CreateTable(model); err != nil {
				return err
			}
		}
	}
	return nil
}

func (db *MysqlDB) CreateDefaultRoles() error {
	defaultRoles := []Role{
		{Name: "admin", Description: "Administrator with full access"},
		{Name: "editor", Description: "Content editor"},
		{Name: "user", Description: "Regular user"},
	}

	for _, role := range defaultRoles {
		var existingRole Role
		result := db.Where("name = ?", role.Name).First(&existingRole)
		if result.Error != nil && errors.Is(result.Error, gorm.ErrRecordNotFound) {
			if err := db.Create(&role).Error; err != nil {
				return fmt.Errorf("failed to create role %s: %w", role.Name, err)
			}
		} else if result.Error != nil {
			return fmt.Errorf("failed to query role %s: %w", role.Name, result.Error)
		}
	}
	return nil
}

// Ping 检查数据库连接是否正常
func (db *MysqlDB) Ping() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// Stats 获取数据库连接统计信息
func (db *MysqlDB) Stats() (map[string]interface{}, error) {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return nil, err
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration,
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}, nil
}
