package orm

import (
	"database/sql"
	"fmt"

	"github.com/MarkRepo/Gee/GeeORM/orm/dialect"
	"github.com/MarkRepo/Gee/GeeORM/orm/log"
	"github.com/MarkRepo/Gee/GeeORM/orm/session"
)

type Engine struct {
	db      *sql.DB
	dialect dialect.Dialect
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

	// make sure the specific dialect exists
	d, ok := dialect.GetDialect(driver)
	if !ok {
		log.Errorf("dialect %s Not Found", driver)
		return nil, fmt.Errorf("error: dialect %s Not Found", driver)
	}

	e := &Engine{db: db, dialect: d}
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
	return session.New(e.db, e.dialect)
}

type TxFunc func(s *session.Session) (interface{}, error)

func (e *Engine) Transaction(f TxFunc) (result interface{}, err error) {
	s := e.NewSession()
	if err := s.Begin(); err != nil {
		return nil, err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = s.Rollback()
			panic(p) // re-throw panic after rollback
		} else if err != nil {
			_ = s.Rollback() // err is non-nil; don't change it
		} else {
			defer func() {
				if err != nil {
					_ = s.Rollback() // if commit return err, rollback
				}
			}()
			err = s.Commit() // err is nil; if Commit returns error update err
		}
	}()

	return f(s)
}
