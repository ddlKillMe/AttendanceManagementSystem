package common

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/spf13/viper"
	"time"
)

// var DB *gorm.DB

var RDS *redis.Pool

//func InitDB() {
//	//drivername := "postgres"
//	host := viper.GetString("datasource.host")
//	user := viper.GetString("datasource.user")
//	password := viper.GetString("datasource.password")
//	dbname := viper.GetString("datasource.dbname")
//	port := viper.GetString("datasource.port")
//	//charset := "utf-8"
//	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
//		host,
//		user,
//		password,
//		dbname,
//		port)
//	//db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
//	db, err := gorm.Open(postgres.New(postgres.Config{
//		DSN:                  dsn,
//		PreferSimpleProtocol: true,
//	}), &gorm.Config{})
//	if err != nil {
//		panic("fail to connect database, err:" + err.Error())
//	}
//	db.AutoMigrate(&model.User{})
//
//	DB = db
//}

//func GetDB() *gorm.DB {
//	return DB
//}

func RedisInit() {
	network := viper.GetString("datasource.network")
	address := viper.GetString("datasource.address")

	RDS = &redis.Pool{
		MaxIdle:     5, //最大空闲数
		MaxActive:   0, //最大连接数，0不设上
		Wait:        true,
		IdleTimeout: time.Duration(1) * time.Second, //空闲等待时间
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial(network, address) //redis IP地址
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			redis.DialDatabase(0)
			return c, err
		},
	}
}

func GetRDS() redis.Conn {
	return RDS.Get()
}

func RedisClose() {
	_ = RDS.Close()
}
