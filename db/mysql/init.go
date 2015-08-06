package mysql

import (
	_ "github.com/go-sql-driver/mysql"

	"time"
	"database/sql"

	"github.com/ottemo/foundation/db"
	"github.com/ottemo/foundation/env"

)

// init makes package self-initialization routine
func init() {
	dbEngine = new(DBEngine)
	dbEngine.attributeTypes = make(map[string]map[string]string)

	var _ db.InterfaceDBEngine = dbEngine

	env.RegisterOnConfigIniStart(dbEngine.Startup)
	db.RegisterDBEngine(dbEngine)
}

// Startup is a database engine startup routines
func (it *DBEngine) Startup() error {

	it.attributeTypes = make(map[string]map[string]string)

	// opening connection
	uri := "/"
	dbName := "ottemo"

	if iniConfig := env.GetIniConfig(); iniConfig != nil {
		if iniValue := iniConfig.GetValue("db.mysql.uri", uri); iniValue != "" {
			uri = iniValue
		}

		if iniValue := iniConfig.GetValue("db.mysql.db", dbName); iniValue != "" {
			dbName = iniValue
		}
	}

	if newConnection, err := sql.Open("mysql", uri); err == nil {
		it.connection = newConnection
	} else {
		return env.ErrorDispatch(err)
	}

	// making sure DB selected otherwise trying to obtain DB
	rows, err := it.connection.Query("SELECT DATABASE()")
	if err != nil {
		return env.ErrorDispatch(err)
	}

	if !rows.Next() || rows.Scan(dbName) != nil || dbName == "" {
		if _, err := it.connection.Exec("USE " + dbName); err != nil {
			if _, err = it.connection.Exec("CREATE DATABASE " + dbName); err != nil {
				return env.ErrorDispatch(err)
			}
			if _, err = it.connection.Exec("USE " + dbName); err != nil {
				return env.ErrorDispatch(err)
			}
		}
	}

	// timer routine to check connection state and reconnect by perforce
	ticker := time.NewTicker(ConstConnectionValidateInterval)
	go func() {
		for _ = range ticker.C {
			err := it.connection.Ping()
			if err != nil {
				dbEngine.connectionMutex.Lock()
				newConnection, err := sql.Open("mysql", uri)
				dbEngine.connectionMutex.Unlock()

				if err != nil {
					env.ErrorDispatch(err)
				} else {
					it.connection = newConnection
				}
			}
		}
	}()

	// making column info table
	SQL := "CREATE TABLE IF NOT EXISTS `" + ConstCollectionNameColumnInfo + "` ( " +
		"`_id`        INTEGER NOT NULL AUTO_INCREMENT," +
		"`collection` VARCHAR(255)," +
		"`column`     VARCHAR(255)," +
		"`type`       VARCHAR(255)," +
		"`indexed`    TINYINT(1)," +
		"PRIMARY KEY(`_id`) )"

	_, err = it.connection.Exec(SQL)
	if err != nil {
		return sqlError(SQL, err)
	}

	db.OnDatabaseStart()

	return nil
}