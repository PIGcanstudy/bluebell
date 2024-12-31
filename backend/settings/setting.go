package settings

import "gopkg.in/ini.v1"

// 从配置文件中读出MySQLConfig的设置
type MySQLConfig struct {
	Host         string `mapstructure:"host"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DB           string `mapstructure:"dbbase"`
	Port         int    `mapstructure:"port"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

// 从配置文件中读出redisConfig的设置
type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	DB           string `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MaxIdleConns int    `mapstructure:"min_idle_conns"`
}

func ReadRedisConfig(cfg *ini.File) RedisConfig {
	return RedisConfig{
		Host:         cfg.Section("redis").Key("host").MustString("localhost"),
		Port:         cfg.Section("redis").Key("port").MustInt(6379),
		Password:     cfg.Section("redis").Key("password").MustString(""),
		DB:           cfg.Section("redis").Key("db").MustString("0"),
		PoolSize:     cfg.Section("redis").Key("pool_size").MustInt(100),
		MaxIdleConns: cfg.Section("redis").Key("max_idle_conns").MustInt(10),
	}
}

func ReadMySQLConfig(cfg *ini.File) MySQLConfig {
	return MySQLConfig{
		Host:         cfg.Section("mysql").Key("host").MustString("localhost"),
		User:         cfg.Section("mysql").Key("user").MustString("root"),
		Password:     cfg.Section("mysql").Key("password").MustString("123456"),
		DB:           cfg.Section("mysql").Key("dbbase").MustString("bluebell"),
		Port:         cfg.Section("mysql").Key("port").MustInt(3306),
		MaxOpenConns: cfg.Section("mysql").Key("max_open_conns").MustInt(100),
		MaxIdleConns: cfg.Section("mysql").Key("max_idle_conns").MustInt(10),
	}
}
