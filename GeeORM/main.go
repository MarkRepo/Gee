package main

import (
	"github.com/MarkRepo/Gee/GeeORM/orm"
	"github.com/MarkRepo/Gee/GeeORM/orm/log"

	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	Name string `geeorm:"PRIMARY KEY"`
	Age  int
}

func main() {
	engine, _ := orm.NewEngine("sqlite3", "gee.db")
	defer engine.Close()
	s := engine.NewSession().Model(&User{})
	_ = s.DropTable()
	_ = s.CreateTable()
	if !s.HasTable() {
		log.Error("Create table failed")
	}
}
