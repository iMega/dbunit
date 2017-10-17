package dbunit

import (
	"database/sql"
	"fmt"
	"testing"
)

type UnitDB interface {
}

// Config is a config for database connection
type Config struct {
	Host   string
	Port   string
	User   string
	DBName string
	Pass   string
}

func (c Config) String() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=true&loc=Local", c.User, c.Pass, c.Host, c.Port, c.DBName)
}

// UnitDB contains information about the state of a database handle
type unitDB struct {
	setup    func(tx *sql.Tx)
	fixtures []func(tx *sql.Tx)
	teardown func(tx *sql.Tx)
	config   string
}

// Option is any options for UnitDB
type Option func(u *unitDB)

// NewUnitDB new instance
func NewUnitDB(t *testing.T, opts ...Option) (UnitDB, func()) {
	u := &unitDB{}
	for _, opt := range opts {
		opt(u)
	}

	db, err := sql.Open("mysql", u.config)
	if err != nil {
		t.Fatalf("Fail to connect to MySQL %s", err)
	}

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Fail to begin transaction %s", err)
	}

	if u.setup != nil {
		u.setup(tx)
	}

	for _, v := range u.fixtures {
		v(tx)
	}

	var teardown = func() {
		if u.teardown != nil {
			u.teardown(tx)
		}
		err = db.Close()
		if err != nil {
			t.Fatalf("Fail closes db connection, %s", err)
		}
	}

	return u, teardown
}

// WithDB sets a config database
func WithConfig(c Config) Option {
	return func(u *unitDB) {
		u.config = c.String()
	}
}

// WithSetup sets initial state a database
func WithSetup(tx func(*sql.Tx)) Option {
	return func(u *unitDB) {
		u.setup = tx
	}
}

// WithFixtures will iterate over all the fixture rows specified and insert them into their respective tables
func WithFixtures(f []func(tx *sql.Tx)) Option {
	return func(u *unitDB) {
		u.fixtures = f
	}
}

// WithTeardown will execute after test
func WithTeardown(tx func(*sql.Tx)) Option {
	return func(u *unitDB) {
		u.teardown = tx
	}
}
