package model

import (
	"database/sql"
	"fmt"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/go-sql-driver/mysql"
	"sync"
)

var (
	mysqlDb   *sql.DB
	mysqlOnce sync.Once
)

type MysqlClient struct {
	db *goqu.Database
}

func NewMysqlClient(dsn string) (*MysqlClient, error) {

	mysqlOnce.Do(func() {
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			return
		}
		mysqlDb = db
	})
	return &MysqlClient{
		db: goqu.New("mysql", mysqlDb),
	}, nil
}

func (that *MysqlClient) AddOne(doc interface{}) (string, error) {
	res, err := that.db.Insert("table_name").Rows(doc).Returning("id").Execute()
	if err != nil {
		return "", err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d", id), nil
}

func (that *MysqlClient) GetOne(id string) (interface{}, error) {
	var result interface{}
	_, err := that.db.From("table_name").Where(goqu.C("id").Eq(id)).ScanStruct(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (that *MysqlClient) UpdateOne(id string, doc interface{}) (interface{}, error) {
	var result interface{}
	_, err := that.db.Update("table_name").Set(doc).Where(goqu.C("id").Eq(id)).Returning("*").ScanStruct(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (that *MysqlClient) DeleteOne(id string) (interface{}, error) {
	var result interface{}
	_, err := that.db.Delete("table_name").Where(goqu.C("id").Eq(id)).Returning("*").ScanStruct(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (that *MysqlClient) Query(filter interface{}) ([]interface{}, error) {
	var results []interface{}
	err := that.db.From("table_name").Where(filter).ScanStructs(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}
