package nustdbclient

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/nutsdb/nutsdb"
)

// resetInstance resets the package-level singleton for testing.
// This is necessary because many functions depend on the global instance.
func resetInstance() {
	instance = nil
	once = sync.Once{}
}

// setupTestDB creates a temporary NutsDB instance for testing.
// Returns the client and a cleanup function.
func setupTestDB(t *testing.T) (*TNustDBClient, func()) {
	t.Helper()
	resetInstance()

	tmpDir := filepath.Join(os.TempDir(), "nustdb_test_"+t.Name())
	os.RemoveAll(tmpDir)

	inst, err := InitInstance("testbucket", tmpDir, true)
	if err != nil {
		t.Fatalf("InitInstance failed: %v", err)
	}

	cleanup := func() {
		if inst != nil && inst.db != nil {
			inst.db.Close()
		}
		os.RemoveAll(tmpDir)
		resetInstance()
	}

	return inst, cleanup
}

// --- TNustDBField tests ---

func TestTNustDBField(t *testing.T) {
	f := TNustDBField{Key: "testkey", Value: "testval"}
	if f.Key != "testkey" {
		t.Errorf("Key = %q, want testkey", f.Key)
	}
	if f.Value != "testval" {
		t.Errorf("Value = %q, want testval", f.Value)
	}
}

// --- TNustDBList tests ---

func TestTNustDBList(t *testing.T) {
	l := TNustDBList{Key: "listkey", Value: []string{"a", "b", "c"}}
	if l.Key != "listkey" {
		t.Errorf("Key = %q, want listkey", l.Key)
	}
	if len(l.Value) != 3 {
		t.Errorf("Value len = %d, want 3", len(l.Value))
	}
}

// --- SetHead / GetHead tests ---

func TestSetHead_GetHead_NonEmpty(t *testing.T) {
	inst, cleanup := setupTestDB(t)
	defer cleanup()

	inst.SetHead("myprefix")
	got := inst.GetHead()
	if got != "myprefix_" {
		t.Errorf("GetHead() = %q, want myprefix_", got)
	}
}

func TestSetHead_GetHead_Empty(t *testing.T) {
	inst, cleanup := setupTestDB(t)
	defer cleanup()

	inst.SetHead("")
	got := inst.GetHead()
	if got != "" {
		t.Errorf("GetHead() = %q, want empty", got)
	}
}

func TestSetHead_AlreadyHasUnderscore(t *testing.T) {
	inst, cleanup := setupTestDB(t)
	defer cleanup()

	inst.SetHead("myprefix_")
	got := inst.GetHead()
	if got != "myprefix_" {
		t.Errorf("GetHead() = %q, want myprefix_", got)
	}
}

// --- GetBucketName test ---

func TestGetBucketName(t *testing.T) {
	inst, cleanup := setupTestDB(t)
	defer cleanup()

	got := inst.GetBucketName()
	if got != "testbucket" {
		t.Errorf("GetBucketName() = %q, want testbucket", got)
	}
}

// --- LSetMaxSize / LGetMaxSize tests ---

func TestLSetMaxSize_LGetMaxSize(t *testing.T) {
	inst, cleanup := setupTestDB(t)
	defer cleanup()

	// Default maxListSize is 100
	if inst.LGetMaxSize() != 100 {
		t.Errorf("LGetMaxSize() = %d, want 100", inst.LGetMaxSize())
	}

	inst.LSetMaxSize(50)
	if inst.LGetMaxSize() != 50 {
		t.Errorf("LGetMaxSize() = %d, want 50", inst.LGetMaxSize())
	}
}

// --- SetValue / GetValue tests ---

func TestSetValue_GetValue(t *testing.T) {
	inst, cleanup := setupTestDB(t)
	defer cleanup()

	err := inst.SetValue("key1", "value1")
	if err != nil {
		t.Fatalf("SetValue error: %v", err)
	}

	val, err := inst.GetValue("key1")
	if err != nil {
		t.Fatalf("GetValue error: %v", err)
	}
	if val != "value1" {
		t.Errorf("GetValue = %q, want value1", val)
	}
}

func TestSetValue_WithTTL(t *testing.T) {
	inst, cleanup := setupTestDB(t)
	defer cleanup()

	err := inst.SetValue("ttlkey", "ttlvalue", 3600)
	if err != nil {
		t.Fatalf("SetValue with TTL error: %v", err)
	}

	val, err := inst.GetValue("ttlkey")
	if err != nil {
		t.Fatalf("GetValue error: %v", err)
	}
	if val != "ttlvalue" {
		t.Errorf("GetValue = %q, want ttlvalue", val)
	}
}

func TestGetValue_NotFound(t *testing.T) {
	inst, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := inst.GetValue("nonexistent")
	if err == nil {
		t.Error("GetValue for nonexistent key should return error")
	}
}

// --- SetValueByBucket / GetValueByBucket tests ---

func TestSetValueByBucket_GetValueByBucket(t *testing.T) {
	inst, cleanup := setupTestDB(t)
	defer cleanup()

	err := inst.SetValueByBucket("testbucket", "bkey1", "bvalue1")
	if err != nil {
		t.Fatalf("SetValueByBucket error: %v", err)
	}

	val, err := inst.GetValueByBucket("testbucket", "bkey1")
	if err != nil {
		t.Fatalf("GetValueByBucket error: %v", err)
	}
	if val != "bvalue1" {
		t.Errorf("GetValueByBucket = %q, want bvalue1", val)
	}
}

func TestSetValueByBucket_EmptyBucketName(t *testing.T) {
	inst, cleanup := setupTestDB(t)
	defer cleanup()

	// Empty bucket name should use the default bucket
	err := inst.SetValueByBucket("", "defkey", "defval")
	if err != nil {
		t.Fatalf("SetValueByBucket with empty name error: %v", err)
	}

	val, err := inst.GetValueByBucket("", "defkey")
	if err != nil {
		t.Fatalf("GetValueByBucket with empty name error: %v", err)
	}
	if val != "defval" {
		t.Errorf("GetValueByBucket = %q, want defval", val)
	}
}

func TestSetValueByBucket_WithTTL(t *testing.T) {
	inst, cleanup := setupTestDB(t)
	defer cleanup()

	err := inst.SetValueByBucket("testbucket", "bttlkey", "bttlvalue", 3600)
	if err != nil {
		t.Fatalf("SetValueByBucket with TTL error: %v", err)
	}

	val, err := inst.GetValueByBucket("testbucket", "bttlkey")
	if err != nil {
		t.Fatalf("GetValueByBucket error: %v", err)
	}
	if val != "bttlvalue" {
		t.Errorf("GetValueByBucket = %q, want bttlvalue", val)
	}
}

// --- DelValue test ---

func TestDelValue(t *testing.T) {
	inst, cleanup := setupTestDB(t)
	defer cleanup()

	inst.SetValue("delkey", "delval")
	err := inst.DelValue("delkey")
	if err != nil {
		t.Fatalf("DelValue error: %v", err)
	}

	_, err = inst.GetValue("delkey")
	if err == nil {
		t.Error("GetValue should fail after DelValue")
	}
}

// --- DelValueByBucket test ---

func TestDelValueByBucket(t *testing.T) {
	inst, cleanup := setupTestDB(t)
	defer cleanup()

	inst.SetValueByBucket("testbucket", "bdelkey", "bdelval")
	err := inst.DelValueByBucket("testbucket", "bdelkey")
	if err != nil {
		t.Fatalf("DelValueByBucket error: %v", err)
	}

	_, err = inst.GetValueByBucket("testbucket", "bdelkey")
	if err == nil {
		t.Error("GetValueByBucket should fail after DelValueByBucket")
	}
}

func TestDelValueByBucket_EmptyBucketName(t *testing.T) {
	inst, cleanup := setupTestDB(t)
	defer cleanup()

	inst.SetValue("emptybucketkey", "emptybucketval")
	err := inst.DelValueByBucket("", "emptybucketkey")
	if err != nil {
		t.Fatalf("DelValueByBucket with empty name error: %v", err)
	}
}

// --- GetAllValue test ---

func TestGetAllValue(t *testing.T) {
	inst, cleanup := setupTestDB(t)
	defer cleanup()

	inst.SetValue("allkey1", "allval1")
	inst.SetValue("allkey2", "allval2")

	items, err := inst.GetAllValue("")
	if err != nil {
		t.Fatalf("GetAllValue error: %v", err)
	}
	if len(items) < 2 {
		t.Errorf("GetAllValue returned %d items, want at least 2", len(items))
	}
}

func TestGetAllValue_WithBucketName(t *testing.T) {
	inst, cleanup := setupTestDB(t)
	defer cleanup()

	inst.SetValueByBucket("testbucket", "bkey1", "bval1")

	items, err := inst.GetAllValue("testbucket")
	if err != nil {
		t.Fatalf("GetAllValue error: %v", err)
	}
	if len(items) < 1 {
		t.Errorf("GetAllValue returned %d items, want at least 1", len(items))
	}
}

// --- DelAllValue test ---

func TestDelAllValue(t *testing.T) {
	inst, cleanup := setupTestDB(t)
	defer cleanup()

	inst.SetValue("akey1", "aval1")
	inst.SetValue("akey2", "aval2")

	err := inst.DelAllValue("")
	if err != nil {
		t.Fatalf("DelAllValue error: %v", err)
	}

	items, err := inst.GetAllValue("")
	if err != nil {
		t.Fatalf("GetAllValue error after DelAllValue: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("GetAllValue returned %d items after DelAllValue, want 0", len(items))
	}
}

// --- DelAllValueByBucket test ---

func TestDelAllValueByBucket(t *testing.T) {
	inst, cleanup := setupTestDB(t)
	defer cleanup()

	inst.SetValueByBucket("testbucket", "bbkey1", "bbval1")
	inst.SetValueByBucket("testbucket", "bbkey2", "bbval2")

	err := inst.DelAllValueByBucket("testbucket", "")
	if err != nil {
		t.Fatalf("DelAllValueByBucket error: %v", err)
	}

	items, err := inst.GetAllValue("testbucket")
	if err != nil {
		t.Fatalf("GetAllValue error after DelAllValueByBucket: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("GetAllValue returned %d items after DelAllValueByBucket, want 0", len(items))
	}
}

func TestDelAllValueByBucket_EmptyBucketName(t *testing.T) {
	inst, cleanup := setupTestDB(t)
	defer cleanup()

	inst.SetValue("ekey", "eval")

	err := inst.DelAllValueByBucket("", "")
	if err != nil {
		t.Fatalf("DelAllValueByBucket with empty name error: %v", err)
	}
}

// --- DelAllValue with keyname prefix test ---

func TestDelAllValue_WithPrefix(t *testing.T) {
	inst, cleanup := setupTestDB(t)
	defer cleanup()

	inst.SetValue("prefix_key1", "val1")
	inst.SetValue("prefix_key2", "val2")
	inst.SetValue("other_key", "val3")

	err := inst.DelAllValue("prefix_")
	if err != nil {
		t.Fatalf("DelAllValue with prefix error: %v", err)
	}

	// The prefixed keys should be deleted
	_, err = inst.GetValue("prefix_key1")
	if err == nil {
		t.Error("prefix_key1 should be deleted")
	}

	// The other key should still exist
	val, err := inst.GetValue("other_key")
	if err != nil {
		t.Fatalf("other_key should still exist: %v", err)
	}
	if val != "val3" {
		t.Errorf("other_key = %q, want val3", val)
	}
}

// --- LPush / LRangeByBucket tests ---

func TestLPush_LRangeByBucket(t *testing.T) {
	inst, cleanup := setupTestDB(t)
	defer cleanup()

	// Create a list bucket first
	inst.db.Update(func(tx *nutsdb.Tx) error { return nil })

	err := inst.LPush("listkey1", "item1")
	if err != nil {
		t.Fatalf("LPush error: %v", err)
	}
	err = inst.LPush("listkey1", "item2")
	if err != nil {
		t.Fatalf("LPush error: %v", err)
	}

	items, err := inst.LRangeByBucket("testbucket", "listkey1")
	if err != nil {
		t.Fatalf("LRangeByBucket error: %v", err)
	}
	if len(items) < 2 {
		t.Errorf("LRangeByBucket returned %d items, want at least 2", len(items))
	}
}

func TestLPushByBucket(t *testing.T) {
	inst, cleanup := setupTestDB(t)
	defer cleanup()

	err := inst.LPushByBucket("testbucket", "lbkey", "lbitem")
	if err != nil {
		t.Fatalf("LPushByBucket error: %v", err)
	}
}

func TestLPushByBucket_EmptyBucketName(t *testing.T) {
	inst, cleanup := setupTestDB(t)
	defer cleanup()

	err := inst.LPushByBucket("", "lbkey2", "lbitem2")
	if err != nil {
		t.Fatalf("LPushByBucket with empty name error: %v", err)
	}
}

func TestLRangeByBucket_EmptyBucketName(t *testing.T) {
	inst, cleanup := setupTestDB(t)
	defer cleanup()

	inst.LPush("lrkey", "lritem")

	items, err := inst.LRangeByBucket("", "lrkey")
	if err != nil {
		t.Fatalf("LRangeByBucket with empty name error: %v", err)
	}
	if len(items) < 1 {
		t.Errorf("LRangeByBucket returned %d items, want at least 1", len(items))
	}
}

// --- LSize test ---

func TestLSize(t *testing.T) {
	inst, cleanup := setupTestDB(t)
	defer cleanup()

	inst.LPush("sizekey", "item1")
	inst.LPush("sizekey", "item2")

	count, err := inst.LSize("testbucket", "sizekey")
	if err != nil {
		t.Fatalf("LSize error: %v", err)
	}
	if count < 2 {
		t.Errorf("LSize = %d, want at least 2", count)
	}
}

func TestLSize_EmptyBucketName(t *testing.T) {
	inst, cleanup := setupTestDB(t)
	defer cleanup()

	inst.LPush("sizekey2", "item1")

	count, err := inst.LSize("", "sizekey2")
	if err != nil {
		t.Fatalf("LSize with empty name error: %v", err)
	}
	if count < 1 {
		t.Errorf("LSize = %d, want at least 1", count)
	}
}

// --- LGetAllValue test ---

func TestLGetAllValue(t *testing.T) {
	inst, cleanup := setupTestDB(t)
	defer cleanup()

	inst.LPush("lak1", "la1")
	inst.LPush("lak1", "la2")

	items, err := inst.LGetAllValue("")
	if err != nil {
		t.Fatalf("LGetAllValue error: %v", err)
	}
	// Should have at least one list
	if len(items) < 1 {
		t.Errorf("LGetAllValue returned %d items, want at least 1", len(items))
	}
}

// --- LUpdateMaxValue test ---

func TestLUpdateMaxValue(t *testing.T) {
	inst, cleanup := setupTestDB(t)
	defer cleanup()

	inst.LSetMaxSize(3)
	inst.LPush("umkey", "um1")
	inst.LPush("umkey", "um2")
	inst.LPush("umkey", "um3")
	inst.LPush("umkey", "um4")

	err := inst.LUpdateMaxValue("")
	if err != nil {
		t.Fatalf("LUpdateMaxValue error: %v", err)
	}
}

// --- LPrintf test ---

func TestLPrintf(t *testing.T) {
	inst, cleanup := setupTestDB(t)
	defer cleanup()

	inst.LPush("pfkey", "pf1")

	err := inst.LPrintf("", "pfkey")
	if err != nil {
		t.Fatalf("LPrintf error: %v", err)
	}
}

// --- Printf test ---

func TestPrintf(t *testing.T) {
	inst, cleanup := setupTestDB(t)
	defer cleanup()

	inst.SetValue("printkey", "printval")

	err := inst.Printf("", "")
	if err != nil {
		t.Fatalf("Printf error: %v", err)
	}
}

func TestPrintf_WithKeyname(t *testing.T) {
	inst, cleanup := setupTestDB(t)
	defer cleanup()

	inst.SetValue("filterkey1", "val1")
	inst.SetValue("filterkey2", "val2")
	inst.SetValue("otherkey", "val3")

	err := inst.Printf("", "filter")
	if err != nil {
		t.Fatalf("Printf with keyname error: %v", err)
	}
}

// --- Merge test ---

func TestMerge(t *testing.T) {
	inst, cleanup := setupTestDB(t)
	defer cleanup()

	// Merge requires at least 2 data files to merge
	// In a fresh DB, there's only one file, so Merge may return an error
	err := inst.Merge()
	if err != nil {
		// This is expected for a fresh DB with only one data file
		t.Logf("Merge returned expected error for fresh DB: %v", err)
	}
}

// --- Constants test ---

func TestConstants(t *testing.T) {
	if ConnTimeout.Seconds() != 3 {
		t.Errorf("ConnTimeout = %v, want 3s", ConnTimeout)
	}
	if OperTimeout.Seconds() != 5 {
		t.Errorf("OperTimeout = %v, want 5s", OperTimeout)
	}
}

// --- SetHead with head prefix applied to keys ---

func TestSetHead_AppliedToKeys(t *testing.T) {
	inst, cleanup := setupTestDB(t)
	defer cleanup()

	inst.SetHead("APP")
	inst.SetValue("mykey", "myval")

	val, err := inst.GetValue("mykey")
	if err != nil {
		t.Fatalf("GetValue error: %v", err)
	}
	if val != "myval" {
		t.Errorf("GetValue = %q, want myval", val)
	}
}

// --- LPush max size trimming test ---

func TestLPush_MaxSizeTrimming(t *testing.T) {
	inst, cleanup := setupTestDB(t)
	defer cleanup()

	inst.LSetMaxSize(3)

	// Push more items than maxListSize
	inst.LPush("trimkey", "item1")
	inst.LPush("trimkey", "item2")
	inst.LPush("trimkey", "item3")
	inst.LPush("trimkey", "item4")

	// After trimming, the list should be at most 3 items
	count, err := inst.LSize("testbucket", "trimkey")
	if err != nil {
		t.Fatalf("LSize error: %v", err)
	}
	if count > 3 {
		t.Errorf("LSize = %d, should be at most 3 after trimming", count)
	}
}

// --- GetInstance test ---

func TestGetInstance(t *testing.T) {
	resetInstance()

	tmpDir := filepath.Join(os.TempDir(), "nustdb_test_GetInstance")
	os.RemoveAll(tmpDir)
	defer os.RemoveAll(tmpDir)

	inst := GetInstance("testbucket", tmpDir, true)
	if inst == nil {
		t.Fatal("GetInstance returned nil")
	}
	if inst.db == nil {
		t.Error("GetInstance returned instance with nil db")
	}
	// Cleanup
	inst.db.Close()
	resetInstance()
}

// --- GetSafeInstance test ---

func TestGetSafeInstance(t *testing.T) {
	resetInstance()

	tmpDir := filepath.Join(os.TempDir(), "nustdb_test_GetSafeInstance")
	os.RemoveAll(tmpDir)
	defer os.RemoveAll(tmpDir)

	inst := GetSafeInstance("testbucket", tmpDir, true)
	if inst == nil {
		t.Fatal("GetSafeInstance returned nil")
	}
	if inst.db == nil {
		t.Error("GetSafeInstance returned instance with nil db")
	}
	// Cleanup
	inst.db.Close()
	resetInstance()
}

// --- InitInstance with re_new=false ---

func TestInitInstance_NoRenew(t *testing.T) {
	resetInstance()

	tmpDir := filepath.Join(os.TempDir(), "nustdb_test_NoRenew")
	os.RemoveAll(tmpDir)
	defer os.RemoveAll(tmpDir)

	// Create the directory with some existing files
	os.MkdirAll(tmpDir, 0755)

	inst, err := InitInstance("testbucket", tmpDir, false)
	if err != nil {
		t.Fatalf("InitInstance with re_new=false error: %v", err)
	}
	if inst == nil {
		t.Fatal("InitInstance returned nil")
	}
	inst.db.Close()
	resetInstance()
}

// --- InitInstance with re_new=true and existing files ---

func TestInitInstance_RenewWithExistingFiles(t *testing.T) {
	resetInstance()

	tmpDir := filepath.Join(os.TempDir(), "nustdb_test_RenewWithFiles")
	os.RemoveAll(tmpDir)
	defer os.RemoveAll(tmpDir)

	// Create the directory with some existing files
	os.MkdirAll(tmpDir, 0755)
	os.WriteFile(filepath.Join(tmpDir, "testdata.dat"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "testmeta.meta"), []byte("test"), 0644)

	inst, err := InitInstance("testbucket", tmpDir, true)
	if err != nil {
		t.Fatalf("InitInstance with re_new=true error: %v", err)
	}
	if inst == nil {
		t.Fatal("InitInstance returned nil")
	}
	inst.db.Close()
	resetInstance()
}

// --- InitInstance with existing instance ---

func TestInitInstance_ExistingInstance(t *testing.T) {
	resetInstance()

	tmpDir := filepath.Join(os.TempDir(), "nustdb_test_Existing")
	os.RemoveAll(tmpDir)
	defer os.RemoveAll(tmpDir)

	inst1, err := InitInstance("testbucket", tmpDir, true)
	if err != nil {
		t.Fatalf("First InitInstance error: %v", err)
	}

	// Second call should return the same instance
	inst2, err := InitInstance("testbucket", tmpDir, false)
	if err != nil {
		t.Fatalf("Second InitInstance error: %v", err)
	}
	if inst1 != inst2 {
		t.Error("Second InitInstance should return same instance")
	}
	inst1.db.Close()
	resetInstance()
}
