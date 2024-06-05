package model

import (
	"database/sql"
	"fmt"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/go-sql-driver/mysql"
	"reflect"
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

func (that *MysqlClient) AddOne(tableName string, doc interface{}) (string, error) {
	res, err := that.db.Insert(tableName).Rows(doc).Returning("id").Executor().Exec()
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

func (that *MysqlClient) UpdateOne(id string, doc interface{}) error {
	_, err := that.db.Update("table_name").Set(doc).Where(goqu.C("id").Eq(id)).Executor().Exec()
	if err != nil {
		return err
	}
	return nil
}

func (that *MysqlClient) DeleteOne(tableName string, id string) error {
	_, err := that.db.Delete(tableName).Where(goqu.C("id").Eq(id)).Executor().Exec()
	if err != nil {
		return err
	}
	return nil
}

func (that *MysqlClient) Query(tableName string, filter interface{}) ([]interface{}, error) {
	var results []interface{}
	gFilter, err := StructToCondition(filter)
	if err != nil {
		return nil, err
	}
	err = that.db.From(tableName).Where(gFilter).ScanStructs(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func StructToRecord(obj interface{}) (goqu.Record, error) {
	// StructToRecord 将结构体转换为 goqu.Record
	record := goqu.Record{}

	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i).Interface()

		dbTag := field.Tag.Get("db")
		if dbTag == "" {
			dbTag = field.Name
		}

		record[dbTag] = value
	}

	return record, nil
}

func StructToCondition(obj interface{}) (goqu.Expression, error) {
	// StructToCondition 将结构体转换为 goqu.C 查询条件
	var conditions []goqu.Expression

	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i).Interface()

		dbTag := field.Tag.Get("db")
		if dbTag == "" {
			dbTag = field.Name
		}

		conditions = append(conditions, goqu.C(dbTag).Eq(value))
	}

	if len(conditions) == 1 {
		return conditions[0], nil
	}

	return goqu.And(conditions...), nil
}
