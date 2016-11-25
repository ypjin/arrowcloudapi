package dao

import (
	"fmt"

	"github.com/astaxie/beego/orm"
	_ "github.com/mattn/go-sqlite3" //register sqlite driver
)

type sqlite struct {
	file string
}

// NewSQLite returns an instance of sqlite
func NewSQLite(file string) Database {
	return &sqlite{
		file: file,
	}
}

// Register registers SQLite as the underlying database used
func (s *sqlite) Register(alias ...string) error {
	if err := orm.RegisterDriver("sqlite3", orm.DRSqlite); err != nil {
		return err
	}

	an := "default"
	if len(alias) != 0 {
		an = alias[0]
	}
	if err := orm.RegisterDataBase(an, "sqlite3", s.file); err != nil {
		return err
	}

	return nil
}

// Name returns the name of SQLite
func (s *sqlite) Name() string {
	return "SQLite"
}

// String returns the details of database
func (s *sqlite) String() string {
	return fmt.Sprintf("type-%s file:%s", s.Name(), s.file)
}
