package util

import (
	"errors"
	"fmt"
	"gin-temp/model"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type MysqlDB struct {
	*gorm.DB
}

func newMysqlDB() *MysqlDB {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		Viper.GetString("mysql.user"),
		Viper.GetString("mysql.password"),
		Viper.GetString("mysql.host"),
		Viper.GetInt("mysql.port"),
		Viper.GetString("mysql.db"),
	)

	db, err := gorm.Open(mysql.New(mysql.Config{DSN: dsn}), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            Viper.GetBool("mysql.prepareStmt"),
		Logger:                 logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		log.Fatal(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}

	maxLifeTime := Viper.GetInt("mysql.maxLifeTime")
	sqlDB.SetMaxOpenConns(Viper.GetInt("mysql.maxOpenConns"))
	sqlDB.SetMaxIdleConns(Viper.GetInt("mysql.maxIdleConns"))
	sqlDB.SetConnMaxLifetime(time.Duration(maxLifeTime) * time.Minute)

	return &MysqlDB{db}
}

func (db *MysqlDB) mysqlMigrate(tables ...interface{}) error {
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

func (db *MysqlDB) addRoot() error {
	if err := db.First(&model.User{}, "username = ?", "root").Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return db.Transaction(func(tx *gorm.DB) error {
				user := model.User{
					Username: "root",
					Password: "root",
					Role:     "root",
				}
				return tx.Save(&user).Error
			})
		}
	}
	return nil
}

func (db *MysqlDB) ping() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

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

func (db *MysqlDB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
