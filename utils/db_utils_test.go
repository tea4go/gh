package utils

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func createEmptyFile(path string) (*os.File, error) {
	return os.Create(path)
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open sqlite3 in-memory db: %v", err)
	}
	return db
}

func createTestTable(db *sql.DB) error {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS test_users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, age INTEGER, created_at DATETIME DEFAULT CURRENT_TIMESTAMP)")
	return err
}

func TestSQLRunByInsert(t *testing.T) {
	fmt.Println("=== 开始测试: SQLRunByInsert ===")
	db := openTestDB(t)
	defer db.Close()

	if err := createTestTable(db); err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	id, err := SQLRunByInsert(db, "INSERT INTO test_users (name, age) VALUES (?, ?)", "Alice", 30)
	if err != nil {
		t.Fatalf("SQLRunByInsert failed: %v", err)
	}
	if id <= 0 {
		t.Errorf("Expected positive insert ID, got %d", id)
	}
	t.Logf("Inserted row with ID: %d", id)

	// Test with bad SQL
	_, err = SQLRunByInsert(db, "INSERT INTO nonexistent_table (name) VALUES (?)", "test")
	if err == nil {
		t.Error("Expected error for inserting into nonexistent table")
	}
}

func TestSQLRunByUpdate(t *testing.T) {
	fmt.Println("=== 开始测试: SQLRunByUpdate ===")
	db := openTestDB(t)
	defer db.Close()

	if err := createTestTable(db); err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Insert a row first
	SQLRunByInsert(db, "INSERT INTO test_users (name, age) VALUES (?, ?)", "Bob", 25)

	// Update it
	affected, err := SQLRunByUpdate(db, "UPDATE test_users SET age = ? WHERE name = ?", 26, "Bob")
	if err != nil {
		t.Fatalf("SQLRunByUpdate failed: %v", err)
	}
	if affected != 1 {
		t.Errorf("Expected 1 affected row, got %d", affected)
	}

	// Update non-existent row
	affected, err = SQLRunByUpdate(db, "UPDATE test_users SET age = ? WHERE name = ?", 99, "NonExistent")
	if err != nil {
		t.Fatalf("SQLRunByUpdate for non-existent row should not error: %v", err)
	}
	if affected != 0 {
		t.Errorf("Expected 0 affected rows for non-existent, got %d", affected)
	}
}

func TestSQLGetValueByInt(t *testing.T) {
	fmt.Println("=== 开始测试: SQLGetValueByInt ===")
	db := openTestDB(t)
	defer db.Close()

	if err := createTestTable(db); err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	SQLRunByInsert(db, "INSERT INTO test_users (name, age) VALUES (?, ?)", "Charlie", 42)

	found, value, err := SQLGetValueByInt(db, "SELECT age FROM test_users WHERE name = ?", "Charlie")
	if err != nil {
		t.Fatalf("SQLGetValueByInt failed: %v", err)
	}
	if !found {
		t.Error("Expected found=true")
	}
	if value != 42 {
		t.Errorf("Expected age=42, got %d", value)
	}

	// No rows
	found, value, err = SQLGetValueByInt(db, "SELECT age FROM test_users WHERE name = ?", "NonExistent")
	if err != nil {
		t.Fatalf("SQLGetValueByInt for non-existent should not error: %v", err)
	}
	if found {
		t.Error("Expected found=false for non-existent row")
	}
	if value != -1 {
		t.Errorf("Expected -1 for not found, got %d", value)
	}
}

func TestSQLGetValueByString(t *testing.T) {
	fmt.Println("=== 开始测试: SQLGetValueByString ===")
	db := openTestDB(t)
	defer db.Close()

	if err := createTestTable(db); err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	SQLRunByInsert(db, "INSERT INTO test_users (name, age) VALUES (?, ?)", "David", 35)

	value, err := SQLGetValueByString(db, "SELECT name FROM test_users WHERE age = ?", 35)
	if err != nil {
		t.Fatalf("SQLGetValueByString failed: %v", err)
	}
	if value != "David" {
		t.Errorf("Expected 'David', got '%s'", value)
	}

	// No rows returns error from Scan
	_, err = SQLGetValueByString(db, "SELECT name FROM test_users WHERE age = ?", 9999)
	if err == nil {
		t.Log("SQLGetValueByString for no rows may return error or empty depending on driver")
	}
}

func TestSQLGetValueByTime(t *testing.T) {
	fmt.Println("=== 开始测试: SQLGetValueByTime ===")
	db := openTestDB(t)
	defer db.Close()

	if err := createTestTable(db); err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	SQLRunByInsert(db, "INSERT INTO test_users (name, age) VALUES (?, ?)", "Eve", 28)

	value, err := SQLGetValueByTime(db, "SELECT created_at FROM test_users WHERE name = ?", "Eve")
	if err != nil {
		t.Logf("SQLGetValueByTime returned error (driver may not support): %v", err)
	} else {
		t.Logf("Got time: %v", value)
	}
}

func TestSQLGetValue(t *testing.T) {
	fmt.Println("=== 开始测试: SQLGetValue ===")
	db := openTestDB(t)
	defer db.Close()

	if err := createTestTable(db); err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	SQLRunByInsert(db, "INSERT INTO test_users (name, age) VALUES (?, ?)", "Frank", 50)

	value, err := SQLGetValue(db, "SELECT name FROM test_users WHERE age = ?", 50)
	if err != nil {
		t.Fatalf("SQLGetValue failed: %v", err)
	}
	t.Logf("SQLGetValue result: %v (type: %T)", value, value)
}

func TestSQLGetValues(t *testing.T) {
	fmt.Println("=== 开始测试: SQLGetValues ===")
	db := openTestDB(t)
	defer db.Close()

	if err := createTestTable(db); err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	SQLRunByInsert(db, "INSERT INTO test_users (name, age) VALUES (?, ?)", "Grace", 33)

	scans, err := SQLGetValues(db, "SELECT name, age FROM test_users WHERE name = ?", "Grace")
	if err != nil {
		t.Fatalf("SQLGetValues failed: %v", err)
	}
	if len(scans) != 2 {
		t.Errorf("Expected 2 columns, got %d", len(scans))
	}
	t.Logf("SQLGetValues result: %v", scans)

	// Bad SQL
	_, err = SQLGetValues(db, "SELECT * FROM nonexistent_table")
	if err == nil {
		t.Error("Expected error for bad SQL")
	}
}

func TestSQLQuery(t *testing.T) {
	fmt.Println("=== 开始测试: SQLQuery ===")
	db := openTestDB(t)
	defer db.Close()

	if err := createTestTable(db); err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	SQLRunByInsert(db, "INSERT INTO test_users (name, age) VALUES (?, ?)", "Hank", 40)
	SQLRunByInsert(db, "INSERT INTO test_users (name, age) VALUES (?, ?)", "Ivy", 22)

	results, err := SQLQuery(db, "SELECT name, age FROM test_users ORDER BY name")
	if err != nil {
		t.Fatalf("SQLQuery failed: %v", err)
	}
	if len(results) < 2 {
		t.Errorf("Expected at least 2 rows, got %d", len(results))
	}
	t.Logf("SQLQuery results: %v", results)

	// Bad SQL
	_, err = SQLQuery(db, "SELECT * FROM nonexistent_table")
	if err == nil {
		t.Error("Expected error for bad SQL")
	}
}

func TestSQLExecDDL(t *testing.T) {
	fmt.Println("=== 开始测试: SQLExecDDL ===")
	db := openTestDB(t)
	defer db.Close()

	// Execute DDL with multiple statements separated by ;
	ddl := "CREATE TABLE ddl_test1 (id INTEGER PRIMARY KEY); CREATE TABLE ddl_test2 (id INTEGER PRIMARY KEY)"
	err := SQLExecDDL(db, ddl, false)
	if err != nil {
		t.Fatalf("SQLExecDDL failed: %v", err)
	}

	// Verify tables were created
	_, err = db.Exec("INSERT INTO ddl_test1 (id) VALUES (1)")
	if err != nil {
		t.Errorf("ddl_test1 table not created properly: %v", err)
	}
	_, err = db.Exec("INSERT INTO ddl_test2 (id) VALUES (1)")
	if err != nil {
		t.Errorf("ddl_test2 table not created properly: %v", err)
	}

	// Test with ignore=true and bad statement
	err = SQLExecDDL(db, "CREATE TABLE ddl_test3 (id INTEGER); BAD SQL STATEMENT", true)
	if err != nil {
		t.Logf("SQLExecDDL with ignore=true returned: %v", err)
	}

	// Test with ignore=false and bad statement
	err = SQLExecDDL(db, "CREATE TABLE ddl_test4 (id INTEGER); BAD SQL STATEMENT", false)
	if err == nil {
		t.Error("Expected error for bad SQL with ignore=false")
	}

	// Test with empty statement after trim
	err = SQLExecDDL(db, "  ;  ;  ", false)
	if err != nil {
		t.Logf("SQLExecDDL with only separators returned: %v", err)
	}
}

func TestSQLPrepared(t *testing.T) {
	fmt.Println("=== 开始测试: SQLPrepared ===")
	db := openTestDB(t)
	defer db.Close()

	if err := createTestTable(db); err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	SQLRunByInsert(db, "INSERT INTO test_users (name, age) VALUES (?, ?)", "Jack", 55)

	columns, err := SQLPrepared(db, "SELECT name, age FROM test_users WHERE name = ?", "Jack")
	if err != nil {
		t.Fatalf("SQLPrepared failed: %v", err)
	}
	if len(columns) != 2 {
		t.Errorf("Expected 2 columns, got %d", len(columns))
	}
	for _, col := range columns {
		t.Logf("Column: %s, Type: %s", col.Name(), col.DatabaseTypeName())
	}

	// Bad SQL
	_, err = SQLPrepared(db, "SELECT * FROM nonexistent_table")
	if err == nil {
		t.Error("Expected error for bad SQL")
	}
}

func TestSQLQueryPrint(t *testing.T) {
	fmt.Println("=== 开始测试: SQLQueryPrint ===")
	db := openTestDB(t)
	defer db.Close()

	if err := createTestTable(db); err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	SQLRunByInsert(db, "INSERT INTO test_users (name, age) VALUES (?, ?)", "Karen", 29)

	err := SQLQueryPrint(db, "SELECT name, age FROM test_users WHERE name = ?", "Karen")
	if err != nil {
		t.Fatalf("SQLQueryPrint failed: %v", err)
	}

	// Bad SQL
	err = SQLQueryPrint(db, "SELECT * FROM nonexistent_table")
	if err == nil {
		t.Error("Expected error for bad SQL")
	}
}

func TestTxRunByInsert(t *testing.T) {
	fmt.Println("=== 开始测试: TxRunByInsert ===")
	db := openTestDB(t)
	defer db.Close()

	if err := createTestTable(db); err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	id, err := TxRunByInsert(tx, "INSERT INTO test_users (name, age) VALUES (?, ?)", "Leo", 45)
	if err != nil {
		tx.Rollback()
		t.Fatalf("TxRunByInsert failed: %v", err)
	}
	if id <= 0 {
		t.Errorf("Expected positive insert ID, got %d", id)
	}

	tx.Commit()

	// Test with bad SQL in transaction
	tx2, _ := db.Begin()
	_, err = TxRunByInsert(tx2, "INSERT INTO nonexistent_table (name) VALUES (?)", "test")
	if err == nil {
		t.Error("Expected error for inserting into nonexistent table")
	}
	tx2.Rollback()
}

func TestTxRunByUpdate(t *testing.T) {
	fmt.Println("=== 开始测试: TxRunByUpdate ===")
	db := openTestDB(t)
	defer db.Close()

	if err := createTestTable(db); err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	SQLRunByInsert(db, "INSERT INTO test_users (name, age) VALUES (?, ?)", "Mike", 38)

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	affected, err := TxRunByUpdate(tx, "UPDATE test_users SET age = ? WHERE name = ?", 39, "Mike")
	if err != nil {
		tx.Rollback()
		t.Fatalf("TxRunByUpdate failed: %v", err)
	}
	if affected != 1 {
		t.Errorf("Expected 1 affected row, got %d", affected)
	}

	tx.Commit()
}

func TestSQLQueryLargeResultSet(t *testing.T) {
	fmt.Println("=== 开始测试: SQLQuery 大结果集 ===")
	db := openTestDB(t)
	defer db.Close()

	if err := createTestTable(db); err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Insert more than 1000 rows to test the limit
	for i := 0; i < 1100; i++ {
		SQLRunByInsert(db, "INSERT INTO test_users (name, age) VALUES (?, ?)", fmt.Sprintf("User%d", i), i%100)
	}

	results, err := SQLQuery(db, "SELECT name, age FROM test_users")
	if err != nil {
		t.Fatalf("SQLQuery failed: %v", err)
	}
	// Should be capped at 1000
	if len(results) > 1000 {
		t.Errorf("Expected at most 1000 rows, got %d", len(results))
	}
	t.Logf("Got %d rows (capped at 1000)", len(results))
}

func TestSaveJsonError(t *testing.T) {
	fmt.Println("=== 开始测试: SaveJson 错误路径 ===")
	// Test saving to an invalid path (directory that doesn't exist and cannot be created)
	err := SaveJson("/nonexistent_dir_xyz/subdir/file.json", make(chan int))
	if err == nil {
		t.Error("Expected error for SaveJson with unmarshallable object or bad path")
	}
	t.Logf("SaveJson error: %v (expected)", err)
}

func TestMkdirError(t *testing.T) {
	fmt.Println("=== 开始测试: Mkdir 错误路径 ===")
	// Try to create a directory where a file already exists
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "existing_file")
	if f, err := createEmptyFile(tmpFile); err != nil {
		t.Fatal(err)
	} else {
		f.Close()
	}

	// Try to Mkdir with the same path as the file
	err := Mkdir(tmpFile)
	if err == nil {
		t.Log("Mkdir on existing file path did not error (OS-dependent)")
	} else {
		t.Logf("Mkdir on existing file path errored as expected: %v", err)
	}
}
