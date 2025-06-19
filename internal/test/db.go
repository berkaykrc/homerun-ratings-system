package test

import (
	"path"
	"regexp"
	"runtime"
	"testing"

	"github.com/berkaykrc/homerun-ratings-system/internal/config"
	"github.com/berkaykrc/homerun-ratings-system/pkg/dbcontext"
	"github.com/berkaykrc/homerun-ratings-system/pkg/log"
	dbx "github.com/go-ozzo/ozzo-dbx"
	_ "github.com/lib/pq" // initialize posgresql for test
)

var db *dbcontext.DB

// DB returns the database connection for testing purpose.
func DB(t *testing.T) *dbcontext.DB {
	if db != nil {
		return db
	}
	logger, _ := log.NewForTest()
	dir := getSourcePath()
	cfg, err := config.Load(dir+"/../../config/local.yml", logger)
	if err != nil {
		t.Fatalf("%v", err)
	}
	dbc, err := dbx.MustOpen("postgres", cfg.DSN)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	dbc.LogFunc = logger.Infof
	db = dbcontext.New(dbc)
	return db
}

// ResetTables truncates all data in the specified tables
func ResetTables(t *testing.T, db *dbcontext.DB, tables ...string) {
	validTableName := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	for _, table := range tables {
		if !validTableName.MatchString(table) {
			t.Fatalf("invalid table name: %s", table)
		}
		// Use CASCADE to handle foreign key constraints automatically
		_, err := db.DB().NewQuery("TRUNCATE TABLE " + table + " RESTART IDENTITY CASCADE").Execute()
		if err != nil {
			t.Fatalf("%v", err)
		}
	}
}

// getSourcePath returns the directory containing the source code that is calling this function.
func getSourcePath() string {
	_, filename, _, _ := runtime.Caller(1)
	return path.Dir(filename)
}
