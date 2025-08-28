package utils

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

func TxRunByUpdate(db_in *sql.Tx, run_sql string, v ...interface{}) (int64, error) {
	stmt, err := db_in.Exec(run_sql, v...)
	if err != nil {
		return 0, err
	}

	update, err := stmt.RowsAffected()

	return update, err
}

func TxRunByInsert(db_in *sql.Tx, run_sql string, v ...interface{}) (int64, error) {
	stmt, err := db_in.Exec(run_sql, v...)
	if err != nil {
		return 0, err
	}

	insert, err := stmt.LastInsertId()

	return insert, err
}

func SQLRunByUpdate(db_in *sql.DB, run_sql string, v ...interface{}) (int64, error) {
	stmt, err := db_in.Exec(run_sql, v...)
	if err != nil {
		return 0, err
	}

	update, err := stmt.RowsAffected()

	return update, err
}

// 执行DDL语句，Oracle驱动不支持执行多个SQL语句，MySQL可以。需要拆分执行。
func SQLExecDDL(db_in *sql.DB, run_sql string, ignore bool) (err error) {
	sql_texts := strings.Split(run_sql, ";")
	if len(sql_texts) == 0 {
		_, err = db_in.Exec(strings.TrimSpace(run_sql))
	} else {
		for k, v := range sql_texts {
			v = strings.TrimSpace(v)
			if v == "" {
				continue
			}
			_, err = db_in.Exec(v)
			if ignore == false && err != nil {
				err = fmt.Errorf("执行第%d条SQL语句失败，%s\n%s", k, v, err.Error())
				break
			}
		}
	}
	return
}

func SQLRunByInsert(db_in *sql.DB, run_sql string, v ...interface{}) (int64, error) {
	stmt, err := db_in.Exec(run_sql, v...)
	if err != nil {
		return 0, err
	}

	insert, err := stmt.LastInsertId()

	return insert, err
}

func SQLGetValueByInt(db_in *sql.DB, query_sql string, v ...interface{}) (bool, int, error) {
	var value int
	err := db_in.QueryRow(query_sql, v...).Scan(&value)
	if err == sql.ErrNoRows {
		return false, -1, nil
	} else {
		return true, value, err
	}
}

func SQLGetValueByTime(db_in *sql.DB, query_sql string, v ...interface{}) (time.Time, error) {
	var value time.Time
	err := db_in.QueryRow(query_sql, v...).Scan(&value)
	return value, err
}

func SQLGetValueByString(db_in *sql.DB, query_sql string, v ...interface{}) (string, error) {
	var value string
	err := db_in.QueryRow(query_sql, v...).Scan(&value)
	return value, err
}

func SQLGetValue(db_in *sql.DB, query_sql string, v ...interface{}) (interface{}, error) {
	value := make([]byte, 0)
	err := db_in.QueryRow(query_sql, v...).Scan(&value)
	return value, err
}

func SQLGetValues(db_in *sql.DB, query_sql string, v ...interface{}) ([]interface{}, error) {
	query, err := db_in.Query(query_sql, v...)
	if err != nil {
		return nil, err
	}
	defer query.Close()

	columns, err := query.Columns()
	if err != nil {
		return nil, err
	}
	values := make([][]byte, len(columns))
	scans := make([]interface{}, len(columns))
	for i := range values { //让每一行数据都填充到[][]byte里面
		scans[i] = &values[i]
	}

	query.Next()
	err = query.Scan(scans...)

	return scans, err
}

func SQLPrepared(db_in *sql.DB, query_sql string, v ...interface{}) ([]*sql.ColumnType, error) {
	query, err := db_in.Query(query_sql, v...)
	if err != nil {
		return nil, err
	}
	defer query.Close()

	columns, err := query.ColumnTypes() //读出查询出的列字段名
	if err != nil {
		return nil, err
	}
	return columns, nil
}

func SQLQuery(db_in *sql.DB, query_sql string, v ...interface{}) ([][]string, error) {
	query, err := db_in.Query(query_sql, v...)
	if err != nil {
		return nil, err
	}
	defer query.Close()

	columns, err := query.Columns() //读出查询出的列字段名
	if err != nil {
		return nil, err
	}

	values := make([][]byte, len(columns))     //values是每个列的值，这里获取到byte里
	scans := make([]interface{}, len(columns)) //因为每次查询出来的列是不定长的，用len(column)定住当次查询的长度
	for i := range values {                    //让每一行数据都填充到[][]byte里面
		scans[i] = &values[i]
	}
	results := make([][]string, 0) //最后得到的map

	//循环，让游标往下移动
	i := 0
	for query.Next() {
		//query.Scan查询出来的不定长值放到scans[i] = &values[i],也就是每行都放在values里
		if err := query.Scan(scans...); err != nil {
			return nil, err
		}
		row := make([]string, len(values)) //每行数据
		for k, v := range values {         //每行数据是放在values里面，现在把它挪到row里
			row[k] = string(v)
		}
		results = append(results, row) //装入结果集中
		i++
		if i >= 1000 {
			break
		}
	}
	return results, nil
}

func SQLQueryPrint(db_in *sql.DB, query_sql string, v ...interface{}) error {
	query, err := db_in.Query(query_sql, v...)
	if err != nil {
		return err
	}
	defer query.Close()
	err = SQLRowPrintByResult(query)
	return err
}

func SQLRowPrintByResult(query *sql.Rows) error {
	column, _ := query.Columns()              //读出查询出的列字段名
	values := make([]string, len(column))     //values是每个列的值，这里获取到byte里
	scans := make([]interface{}, len(column)) //因为每次查询出来的列是不定长的，用len(column)定住当次查询的长度
	for i := range values {                   //让每一行数据都填充到[][]byte里面
		scans[i] = &values[i]
	}

	results := make(map[int]map[string]string) //最后得到的map
	i := 0

	//循环，让游标往下移动
	for query.Next() {
		//query.Scan查询出来的不定长值放到scans[i] = &values[i],也就是每行都放在values里
		if err := query.Scan(scans...); err != nil {
			return err
		}
		row := make(map[string]string) //每行数据
		for k, v := range values {     //每行数据是放在values里面，现在把它挪到row里
			key := column[k]
			fmt.Println(key, string(v))
			row[key] = string(v)
		}
		results[i] = row //装入结果集中
		i++
		if i >= 2 {
			break
		}
	}
	for k, v := range results { //查询出来的数组
		fmt.Println(k, v)
	}
	return nil
}
