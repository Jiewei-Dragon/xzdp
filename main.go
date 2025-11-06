package main

import (
	"log/slog"
	"xzdp/config"
	"xzdp/db"
	"xzdp/pkg/logger"
	"xzdp/router"

	"github.com/spf13/pflag"
)

func init() {
	configPath := pflag.StringP("config", "c", "configs/config.yaml", "config file path")
	pflag.Parse()

	config.InitConfig(*configPath)      //初始化配置
	logger.InitLogger(config.LogOption) //初始化日志
	slog.Info("MySQL和Redis配置成功")
	var err error
	//初始化数据库
	db.DBEngine, err = db.NewMySQL(config.MysqlOption)
	if err != nil {
		panic(err)
	}
	db.RedisDb, err = db.NewRedisClient(config.RedisOption)
	if err != nil {
		panic(err)
	}
}

func main() {
	r := router.NewRouter()
	err := r.Run(":" + config.ServerOption.HttpPort)
	if err != nil {
		panic(err)
	}
}
