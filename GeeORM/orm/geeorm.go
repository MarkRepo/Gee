package orm

import (
	"database/sql"

	"github.com/MarkRepo/Gee/GeeORM/orm/log"
	"github.com/MarkRepo/Gee/GeeORM/orm/session"
)

type Engine struct {
	db *sql.DB
}

// NewEngine connect database
func NewEngine(driver, database string) (*Engine, error) {
	db, err := sql.Open(driver, database)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	if err := db.Ping(); err != nil {
		log.Error(err)
		return nil, err
	}

	e := &Engine{db: db}
	log.Infof("Connect database %s success", database)
	return e, nil
}

func (e *Engine) Close() {
	if err := e.db.Close(); err != nil {
		log.Errorf("Failed to close database error: %v", err)
	}
	log.Info("close database success")
}

func (e *Engine) NewSession() *session.Session {
	return session.New(e.db)
}
