// Package dialect 隔离各个数据库之间的差异
package dialect

import "reflect"

var dialectsMap = map[string]Dialect{}

// Dialect 各个数据库需要实现的接口
//go:generate mockery --name Dialect --case snake
type Dialect interface {
	// DataTypeOf 用于将 Go 语言的类型转换为该数据库的数据类型
	DataTypeOf(typ reflect.Value) string
	// TableExistSQL 返回某个表是否存在的 SQL 语句
	TableExistSQL(tableName string) (string, []interface{})
}

// RegisterDialect 注册数据库实现实例
func RegisterDialect(name string, dialect Dialect) {
	dialectsMap[name] = dialect
}

// GetDialect 获取 Dialect实例
func GetDialect(name string) (dialect Dialect, ok bool) {
	dialect, ok = dialectsMap[name]
	return
}
