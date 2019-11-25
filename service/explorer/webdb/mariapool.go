package webdb

import (

	// mysql driver
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"

	// gorm mysql driver
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// Connect DB
func Connect() *gorm.DB {
	// db, err := gorm.Open("mysql", "portal_admin:nhcq3t40n9y32t4f0n8@tcp(49.247.203.232)/f_portal?charset=utf8mb4&parseTime=True&loc=Local")
	db, err := gorm.Open("mysql", "portal_admin:nhcq3t40n9y32t4f0n8@tcp(127.0.0.1)/f_portal?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		// panic("failed to connect database")
		panic(err)
	}
	db.DB().SetMaxIdleConns(4)
	db.DB().SetMaxOpenConns(8)

	return db
}
