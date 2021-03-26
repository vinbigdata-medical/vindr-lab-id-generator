package generator

import (
	"fmt"
	"time"

	"github.com/rqlite/gorqlite"

	"go.uber.org/zap"
)

type GeneratorDB struct {
	conn   *gorqlite.Connection
	logger *zap.Logger
}

func NewGeneratorStore(conn *gorqlite.Connection, logger *zap.Logger) *GeneratorDB {
	return &GeneratorDB{conn: conn, logger: logger}
}

func (db *GeneratorDB) CreateTableIfNotExist(tableName string) error {
	createCounterTableSQL := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s ("seed" integer NOT NULL PRIMARY KEY AUTOINCREMENT,"modified" INTEGER);`, tableName) // SQL Statement for Create Table
	_, err := db.conn.WriteOne(createCounterTableSQL)
	if err != nil {
		db.logger.Error(err.Error())
		return err
	}
	return nil
}

func (db *GeneratorDB) Set(tableName string, value int64) (int64, error) {
	now := time.Now().UnixNano() / int64(time.Millisecond)
	insertCounter := fmt.Sprintf(`INSERT INTO %s (seed,modified) VALUES (%d,%d)`, tableName, value, now)
	res, err := db.conn.WriteOne(insertCounter)
	if err != nil {
		db.logger.Error(err.Error())
		return -1, err
	}
	seed := res.LastInsertID
	return seed, nil
}

func (db *GeneratorDB) Insert(tableName string) (int64, error) {
	now := time.Now().UnixNano() / int64(time.Millisecond)
	insertCounter := fmt.Sprintf(`INSERT INTO %s (modified) VALUES (%d)`, tableName, now)
	res, err := db.conn.WriteOne(insertCounter)
	if err != nil {
		db.logger.Error(err.Error())
		return -1, err
	}
	seed := res.LastInsertID
	return seed, nil
}

func (db *GeneratorDB) QueryLast(tableName string) (int, error) {
	queryCounter := fmt.Sprintf(`SELECT seed FROM %s ORDER BY seed DESC`, tableName)
	rows, err := db.conn.QueryOne(queryCounter)
	if err != nil {
		db.logger.Error(err.Error())
		return -1, err
	}

	seed := 0
	for rows.Next() {
		err := rows.Scan(&seed)
		if err != nil {
			db.logger.Error(err.Error())
			return -1, err
		}
		break
	}
	return seed, nil
}

func (db *GeneratorDB) Delete(tableName string, upperBound int64) error {
	queryCounter := fmt.Sprintf(`DELETE FROM %s	WHERE seed < %d`, tableName, upperBound)
	_, err := db.conn.WriteOne(queryCounter)
	if err != nil {
		db.logger.Error(err.Error())
		return err
	}
	return nil
}

func (db *GeneratorDB) ShowAllTables() error {
	queryCounter := fmt.Sprintf(`SELECT name FROM sqlite_master WHERE type='table'`)
	_, err := db.conn.QueryOne(queryCounter)
	if err != nil {
		db.logger.Error(err.Error())
		return err
	}
	return nil
}
