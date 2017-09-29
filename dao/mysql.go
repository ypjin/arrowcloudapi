package dao

import (
	"errors"
	"fmt"
	"net"

	"time"

	"arrowcloudapi/utils/log"

	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql" //register mysql driver
)

type mysql struct {
	host     string
	port     string
	usr      string
	pwd      string
	database string
}

// NewMySQL returns an instance of mysql
func NewMySQL(host, port, usr, pwd, database string) Database {
	return &mysql{
		host:     host,
		port:     port,
		usr:      usr,
		pwd:      pwd,
		database: database,
	}
}

// Register registers MySQL as the underlying database used
func (m *mysql) Register(alias ...string) error {
	if err := m.testConn(m.host, m.port); err != nil {
		return err
	}

	if err := orm.RegisterDriver("mysql", orm.DRMySQL); err != nil {
		return err
	}

	an := "default"
	if len(alias) != 0 {
		an = alias[0]
	}
	conn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", m.usr,
		m.pwd, m.host, m.port, m.database)
	return orm.RegisterDataBase(an, "mysql", conn)
}

func (m *mysql) testConn(host, port string) error {
	ch := make(chan int, 1)
	go func() {
		var err error
		var c net.Conn
		for {
			c, err = net.DialTimeout("tcp", host+":"+port, 20*time.Second)
			if err == nil {
				c.Close()
				ch <- 1
			} else {
				log.Errorf("failed to connect to db, retry after 2 seconds :%v", err)
				time.Sleep(2 * time.Second)
			}
		}
	}()
	select {
	case <-ch:
		return nil
	case <-time.After(60 * time.Second):
		return errors.New("failed to connect to database after 60 seconds")
	}
}

// Name returns the name of MySQL
func (m *mysql) Name() string {
	return "MySQL"
}

// String returns the details of database
func (m *mysql) String() string {
	return fmt.Sprintf("type-%s host-%s port-%s user-%s database-%s",
		m.Name(), m.host, m.port, m.usr, m.database)
}
