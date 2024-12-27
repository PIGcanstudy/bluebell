package mysql

import (
	"fmt"

	"gopkg.in/ini.v1"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

// 此会执行在main函数之前
func Init(cfg *ini.File) error {
	// "root:123456@(localhost:3306)/bluebell?charset=utf8mb4&parseTime=True&loc=Local"
	dsn := fmt.Sprintf("%s:%s@(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", cfg.Section("mysql").Key("user").String(), cfg.Section("mysql").Key("password").String(),
		cfg.Section("mysql").Key("host").String(), cfg.Section("mysql").Key("port").String(), cfg.Section("mysql").Key("database").String())

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		return err
	}

	sqlDB, err := DB.DB()

	if err != nil {
		return err
	}

	sqlDB.SetMaxOpenConns(cfg.Section("mysql").Key("max_open_conns").MustInt(200))

	sqlDB.SetMaxIdleConns(cfg.Section("mysql").Key("max_idle_conns").MustInt(50))

	return nil
}
