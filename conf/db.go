package conf

import (
	"github.com/samber/lo"
	"time"
)

const (
	MySQL   DBDriver       = "mysql"
	Source  DBResolverType = "source"
	Replica DBResolverType = "replica"
)

var supportedDrivers = []DBDriver{MySQL}

type DBDriver string
type DBResolverType string

// DB 数据库配置结构体
type DB struct {
	Driver          DBDriver       `json:"driver" yaml:"driver"`
	Type            DBResolverType `json:"-" yaml:"-"`
	Addr            string         `json:"addr" yaml:"addr"`
	Database        string         `json:"database" yaml:"database"`
	Username        string         `json:"username" yaml:"username"`
	Password        string         `json:"password" yaml:"password"`
	Options         string         `json:"options" yaml:"options"`
	MaxDialTimeout  time.Duration  `json:"maxDialTimeout" yaml:"maxDialTimeout"`
	MaxIdleConn     int            `json:"maxIdleConn" yaml:"maxIdleConn"`
	MaxOpenConn     int            `json:"maxOpenConn" yaml:"maxOpenConn"`
	ConnMaxIdleTime time.Duration  `json:"connMaxIdleTime" yaml:"connMaxIdleTime"`
	ConnMaxLifeTime time.Duration  `json:"connMaxLifeTime" yaml:"connMaxLifeTime"`
	LogInfo         bool           `json:"logInfo" yaml:"logInfo"`
	Resolvers       []*DBResolver  `json:"resolvers" yaml:"resolvers"`
}

// DBConn 数据库连接配置结构体
type DBConn struct {
	Driver          DBDriver      `json:"driver" yaml:"driver"`
	Addr            string        `json:"addr" yaml:"addr"`
	Database        string        `json:"database" yaml:"database"`
	Username        string        `json:"username" yaml:"username"`
	Password        string        `json:"password" yaml:"password"`
	Options         string        `json:"options" yaml:"options"`
	MaxDialTimeout  time.Duration `json:"maxDialTimeout" yaml:"maxDialTimeout"`
	MaxIdleConn     int           `json:"maxIdleConn" yaml:"maxIdleConn"`
	MaxOpenConn     int           `json:"maxOpenConn" yaml:"maxOpenConn"`
	ConnMaxIdleTime time.Duration `json:"connMaxIdleTime" yaml:"connMaxIdleTime"`
	ConnMaxLifeTime time.Duration `json:"connMaxLifeTime" yaml:"connMaxLifeTime"`
}

// DbDsn 数据库DBS配置
type DbDsn struct {
	Driver   DBDriver `json:"driver" yaml:"driver"`
	Addr     string   `json:"addr" yaml:"addr"`
	Database string   `json:"database" yaml:"database"`
	Username string   `json:"username" yaml:"username"`
	Password string   `json:"password" yaml:"password"`
	Options  string   `json:"options" yaml:"options"`
}

// DBConnPool 数据库连接池配置
type DBConnPool struct {
	MaxIdleConn     int           `json:"maxIdleConn" yaml:"maxIdleConn"`
	MaxOpenConn     int           `json:"maxOpenConn" yaml:"maxOpenConn"`
	ConnMaxIdleTime time.Duration `json:"connMaxIdleTime" yaml:"connMaxIdleTime"`
	ConnMaxLifeTime time.Duration `json:"connMaxLifeTime" yaml:"connMaxLifeTime"`
}

// DBResolver 数据库主从配置
type DBResolver struct {
	Driver          DBDriver       `json:"-" yaml:"-"`
	Type            DBResolverType `json:"type" yaml:"type"`
	Addr            string         `json:"addr" yaml:"addr"`
	Database        string         `json:"database" yaml:"database"`
	Username        string         `json:"username" yaml:"username"`
	Password        string         `json:"password" yaml:"password"`
	Options         string         `json:"options" yaml:"options"`
	MaxIdleConn     int            `json:"maxIdleConn" yaml:"maxIdleConn"`
	MaxOpenConn     int            `json:"maxOpenConn" yaml:"maxOpenConn"`
	ConnMaxIdleTime time.Duration  `json:"connMaxIdleTime" yaml:"connMaxIdleTime"`
	ConnMaxLifeTime time.Duration  `json:"connMaxLifeTime" yaml:"connMaxLifeTime"`
}

// ToString String 转换为字符串
func (d DBDriver) ToString() string {
	return string(d)
}

func (r DBResolverType) ToString() string {
	return string(r)
}

// IsSupported 检查是否支持的驱动
func (d DBDriver) IsSupported() bool {
	return lo.Contains(supportedDrivers, d)
}

func (db DB) Equals(other DB) bool {
	return db.Addr == other.Addr &&
		db.Driver == other.Driver &&
		db.Type == other.Type &&
		db.Database == other.Database &&
		db.Username == other.Username &&
		db.Password == other.Password &&
		db.Options == other.Options &&
		db.MaxIdleConn == other.MaxIdleConn &&
		db.MaxOpenConn == other.MaxOpenConn &&
		db.ConnMaxIdleTime == other.ConnMaxIdleTime &&
		db.ConnMaxLifeTime == other.ConnMaxLifeTime &&
		db.LogInfo == other.LogInfo
}

func (db DB) NotEquals(other DB) bool {
	return !db.Equals(other)
}
