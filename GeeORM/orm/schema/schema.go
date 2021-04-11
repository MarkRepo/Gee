// Package schema 实现对象和表的转化
// 表名(table name)	—— 结构体名(struct name)
// 字段名和字段类型	—— 成员变量和类型。
// 额外的约束条件(例如非空、主键等) —— 成员变量的Tag（Go 语言通过 Tag 实现，Java、Python 等语言通过注解实现）
package schema

import (
	"go/ast"
	"reflect"

	"github.com/MarkRepo/Gee/GeeORM/orm/dialect"
)

// Field represents a column of database
type Field struct {
	Name string
	Type string
	Tag  string
}

// Schema represents a table of database
type Schema struct {
	Model      interface{} // Model 被映射的对象
	Name       string      // Name 表名
	Fields     []*Field    // Fields 字段
	FieldNames []string
	FiledMap   map[string]*Field
}

func (s *Schema) GetField(name string) *Field {
	return s.FiledMap[name]
}

// Parse 将任意的对象解析为 Schema 实例
func Parse(dest interface{}, d dialect.Dialect) *Schema {
	modelType := reflect.Indirect(reflect.ValueOf(dest)).Type()
	schema := &Schema{
		Model:    dest,
		Name:     modelType.Name(),
		FiledMap: make(map[string]*Field),
	}

	for i := 0; i < modelType.NumField(); i++ {
		p := modelType.Field(i)
		if !p.Anonymous && ast.IsExported(p.Name) {
			field := &Field{
				Name: p.Name,
				Type: d.DataTypeOf(reflect.Indirect(reflect.New(p.Type))),
			}
			if v, ok := p.Tag.Lookup("geeorm"); ok {
				field.Tag = v
			}
			schema.Fields = append(schema.Fields, field)
			schema.FieldNames = append(schema.FieldNames, field.Name)
			schema.FiledMap[field.Name] = field
		}
	}

	return schema
}
