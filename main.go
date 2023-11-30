package main

import (
	"dning.com/pro02/common"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"os"
)

func main() {
	InitConfig2()
	common.RedisInit()
	defer common.RedisClose()

	r := gin.Default()
	r = CollectRoute(r)
	port := viper.GetString("server.port")
	if port != "" {
		panic(r.Run(":" + port))
	}

	panic(r.Run())
}

func InitConfig() {
	workdir, _ := os.Getwd()
	viper.SetConfigName("application")
	viper.SetConfigType("json")
	viper.AddConfigPath(workdir + "/config")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

func InitConfig2() {
	workdir, _ := os.Getwd()
	viper.SetConfigName("application2")
	viper.SetConfigType("json")
	viper.AddConfigPath(workdir + "/config")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}
