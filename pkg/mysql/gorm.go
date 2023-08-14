package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/stubborn-gaga-0805/aurora/conf"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
	"time"
)

var ErrUnsupportedDriver = errors.New("unsupported database driver")
var ErrUnsupportedResolverType = errors.New("unsupported resolver type")

func New(ctx context.Context, c conf.DB) (gdb *gorm.DB, err error) {

	sdb, err := initDb(ctx, c)
	if err != nil {
		return nil, err
	}

	dialect, err := NewDialect(c.Driver, sdb)
	if err != nil {
		return nil, err
	}

	gdb, err = gorm.Open(dialect, &gorm.Config{
		SkipDefaultTransaction:                   true,
		DisableForeignKeyConstraintWhenMigrating: true,
	})

	if err = registerResolver(ctx, gdb, c); err != nil {
		return nil, err
	}

	return gdb, nil
}

func registerResolver(ctx context.Context, gdb *gorm.DB, c conf.DB) error {
	rcs := c.Resolvers

	resolvers := make([]conf.DB, 0, len(rcs))
	for _, rc := range rcs {
		var dbRes = conf.DB{
			Driver:   c.Driver,
			Type:     rc.Type,
			Addr:     rc.Addr,
			Database: rc.Database,
			Username: rc.Username,
			Password: rc.Password,
			Options:  rc.Options,
		}

		if rc.MaxIdleConn == 0 {
			dbRes.MaxIdleConn = c.MaxIdleConn
		}
		if rc.MaxOpenConn == 0 {
			dbRes.MaxOpenConn = c.MaxOpenConn
		}
		if rc.ConnMaxIdleTime == 0 {
			dbRes.ConnMaxIdleTime = c.ConnMaxIdleTime
		}
		if rc.ConnMaxLifeTime == 0 {
			dbRes.ConnMaxLifeTime = c.ConnMaxLifeTime
		}

		resolvers = append(resolvers, dbRes)
	}

	plugin, err := buildResolver(ctx, resolvers)
	if err != nil {
		return err
	}
	return gdb.Use(plugin)
}

func buildResolver(ctx context.Context, resolvers []conf.DB) (gorm.Plugin, error) {
	var (
		sources  = make([]gorm.Dialector, 0, len(resolvers))
		replicas = make([]gorm.Dialector, 0, len(resolvers))
	)

	for _, resolver := range resolvers {
		sdb, err := initDb(ctx, resolver)
		if err != nil {
			return nil, err
		}

		dialect, err := NewDialect(resolver.Driver, sdb)
		if err != nil {
			return nil, err
		}

		switch resolver.Type {
		case conf.Source:
			sources = append(sources, dialect)
		case conf.Replica:
			replicas = append(replicas, dialect)
		default:
			return nil, ErrUnsupportedResolverType
		}
	}

	return dbresolver.Register(dbresolver.Config{
		Sources:  sources,
		Replicas: replicas,
		Policy:   dbresolver.RandomPolicy{},
	}), nil
}

func buildDSN(c conf.DB) string {
	dsn := ""
	dialTimeOut := time.Second * 2
	if c.MaxDialTimeout.Seconds() > 0 {
		dialTimeOut = c.MaxDialTimeout
	}

	switch c.Driver {
	case conf.MySQL:
		fallthrough
	default:
		dsn = fmt.Sprintf("%s:%s@tcp(%s)/%s?%s&timeout=%s", c.Username, c.Password, c.Addr, c.Database, c.Options, dialTimeOut.String())
	}

	return dsn
}

func initDb(ctx context.Context, c conf.DB) (*sql.DB, error) {
	if !c.Driver.IsSupported() {
		return nil, ErrUnsupportedDriver
	}

	dsn := buildDSN(c)

	db, err := sql.Open(c.Driver.ToString(), dsn)
	if err != nil {
		return nil, err
	}

	if c.MaxIdleConn > 0 {
		db.SetMaxIdleConns(c.MaxIdleConn)
	}

	if c.MaxOpenConn > 0 {
		db.SetMaxOpenConns(c.MaxOpenConn)
	}

	if c.ConnMaxIdleTime > 0 {
		db.SetConnMaxLifetime(c.ConnMaxIdleTime)
	}

	if c.ConnMaxLifeTime > 0 {
		db.SetConnMaxLifetime(c.ConnMaxLifeTime)
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}
