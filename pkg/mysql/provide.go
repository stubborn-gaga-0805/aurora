package mysql

import (
	"context"
	"github.com/stubborn-gaga-0805/aurora/conf"
	"gorm.io/gorm"
)

func NewMysql(ctx context.Context, c conf.DB) *gorm.DB {
	db, err := New(ctx, c)
	if err != nil {
		panic(err)
	}

	return db
}
