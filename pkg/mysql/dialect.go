package mysql

import (
	"github.com/aurora/conf"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewDialect(driver conf.DBDriver, conn gorm.ConnPool) (gorm.Dialector, error) {
	var dialect gorm.Dialector
	switch driver {
	case conf.MySQL:
		dialect = newMySQLDialect(conn)
	default:
		return nil, ErrUnsupportedDriver
	}
	return dialect, nil
}

// newMySQLDialect build mysql dialect
func newMySQLDialect(conn gorm.ConnPool) gorm.Dialector {
	return mysql.New(mysql.Config{Conn: conn})
}
