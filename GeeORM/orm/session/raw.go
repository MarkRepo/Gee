// Package session 负责与数据库交互的封装
package session

import (
	"database/sql"
	"strings"

	// internal
	"github.com/MarkRepo/Gee/GeeORM/orm/clause"
	"github.com/MarkRepo/Gee/GeeORM/orm/dialect"
	"github.com/MarkRepo/Gee/GeeORM/orm/log"
	"github.com/MarkRepo/Gee/GeeORM/orm/schema"
)

// CommonDB is a minimal function set of db
type CommonDB interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// Session op to db
type Session struct {
	db       *sql.DB
	tx       *sql.Tx // 当 tx 不为空时，则使用 tx 执行 SQL 语句，否则使用 db 执行 SQL 语句
	dialect  dialect.Dialect
	refTable *schema.Schema
	clause   clause.Clause
	sql      strings.Builder
	sqlVars  []interface{}
}

// New create a new session
func New(db *sql.DB, d dialect.Dialect) *Session {
	return &Session{db: db, dialect: d}
}

// Clear reset sql and sqlVars
func (s *Session) Clear() {
	s.sql.Reset()
	s.sqlVars = nil
	s.clause = clause.Clause{}
}

// DB return db
func (s *Session) DB() CommonDB {
	if s.tx != nil {
		return s.tx
	}
	return s.db
}

// Raw set sql and sqlVars
func (s *Session) Raw(sql string, values ...interface{}) *Session {
	s.sql.WriteString(sql)
	s.sql.WriteString(" ")
	s.sqlVars = append(s.sqlVars, values...)
	return s
}

// Exec with sql and sqlVars
func (s *Session) Exec() (result sql.Result, err error) {
	defer s.Clear()
	log.Info(s.sql.String(), s.sqlVars)
	if result, err = s.DB().Exec(s.sql.String(), s.sqlVars...); err != nil {
		log.Error(err)
	}
	return
}

// QueryRow get a record from db
func (s *Session) QueryRow() *sql.Row {
	defer s.Clear()
	log.Info(s.sql.String(), s.sqlVars)
	return s.DB().QueryRow(s.sql.String(), s.sqlVars...)
}

// QueryRows get records from db
func (s *Session) QueryRows() (rows *sql.Rows, err error) {
	defer s.Clear()
	log.Info(s.sql.String(), s.sqlVars)
	if rows, err = s.DB().Query(s.sql.String(), s.sqlVars...); err != nil {
		log.Error(err)
	}
	return
}
