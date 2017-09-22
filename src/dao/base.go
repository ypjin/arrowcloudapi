package dao

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"sync"

	"utils/log"

	"github.com/astaxie/beego/orm"
)

// NonExistUserID : if a user does not exist, the ID of the user will be 0.
const NonExistUserID = 0

// Database is an interface of different databases
type Database interface {
	// Name returns the name of database
	Name() string
	// String returns the details of database
	String() string
	// Register registers the database which will be used
	Register(alias ...string) error
}

// InitDatabase initializes the database
func InitDatabase() {
	database, err := getDatabase()
	if err != nil {
		panic(err)
	}

	log.Infof("initializing database: %s", database.String())
	if err := database.Register(); err != nil {
		panic(err)
	}
}

func getDatabase() (db Database, err error) {
	switch strings.ToLower(os.Getenv("DATABASE")) {
	case "", "mysql":
		host, port, usr, pwd, database := getMySQLConnInfo()
		db = NewMySQL(host, port, usr, pwd, database)
	case "sqlite":
		file := getSQLiteConnInfo()
		db = NewSQLite(file)
	default:
		err = fmt.Errorf("invalid database: %s", os.Getenv("DATABASE"))
	}

	return
}

// TODO read from config
func getMySQLConnInfo() (host, port, username, password, database string) {
	host = os.Getenv("MYSQL_HOST")
	port = os.Getenv("MYSQL_PORT")
	username = os.Getenv("MYSQL_USR")
	password = os.Getenv("MYSQL_PWD")
	database = os.Getenv("MYSQL_DATABASE")
	if len(database) == 0 {
		database = "registry"
	}
	return
}

// TODO read from config
func getSQLiteConnInfo() string {
	file := os.Getenv("SQLITE_FILE")
	if len(file) == 0 {
		file = "registry.db"
	}
	return file
}

var globalOrm orm.Ormer
var once sync.Once

// GetOrmer :set ormer singleton
func GetOrmer() orm.Ormer {
	debug.PrintStack()
	once.Do(func() {
		globalOrm = orm.NewOrm()
	})
	return globalOrm
}

func paginateForRawSQL(sql string, limit, offset int64) string {
	return fmt.Sprintf("%s limit %d offset %d", sql, limit, offset)
}
