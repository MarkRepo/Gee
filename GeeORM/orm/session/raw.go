// Package session 负责与数据库交互的封装
package session

import (
	"database/sql"
	"strings"

	// internal
	"github.com/MarkRepo/Gee/GeeORM/orm/log"
)

// Session op to db
type Session struct {
	db      *sql.DB
	sql     strings.Builder
	sqlVars []interface{}
}

// New create a new session
func New(db *sql.DB) *Session {
	return &Session{db: db}
}

// Clear reset sql and sqlVars
func (s *Session) Clear() {
	s.sql.Reset()
	s.sqlVars = nil
}

// DB return db
func (s *Session) DB() *sql.DB {
	return s.db
}

// Raw set sql and sqlVars
func (s *Session) Raw(sql string, values ...interface{}) *Session {
	s.sql.WriteString(sql)
	s.sql.WriteString(" ")
	s.sqlVars = append(s.sqlVars, values...)
	return s
}

// Exec raw sql with sqlVars
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
