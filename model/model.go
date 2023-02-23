package model

import (
	"fmt"
	"log"

	"github.com/jary-287/gopass-common/common"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var Db *gorm.DB

func GetDB() error {
	c, err := common.GetConsulConfig("192.168.0.19", 8500, "micro/config")
	if err != nil {
		return err
	}
	mysqlInfo := common.GetMysqlFromConsul(c, "micro", "config", "mysql")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		mysqlInfo.User,
		mysqlInfo.PassWord,
		mysqlInfo.Host,
		mysqlInfo.Port,
		mysqlInfo.Database)
	Db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction:                   true, //禁止默认事务
		DisableForeignKeyConstraintWhenMigrating: true, //关闭创建外键约束
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, //禁止复表
		},
	})

	if err != nil {
		return err
	}
	log.Println("mysql connection success")
	sqlDB, _ := Db.DB()
	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(10)
	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(100)
	return nil
}
