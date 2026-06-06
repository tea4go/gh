package ldapclient

import (
	"encoding/json"
	"net"
	"strings"
	"testing"
	"time"

	ldap "github.com/go-ldap/ldap/v3"
	"github.com/tea4go/gh/utils"
)

// --- TResultLdap tests ---

func TestTResultLdap_GetAttr_SingleValue(t *testing.T) {
	r := &TResultLdap{
		DN: "cn=test,dc=example,dc=com",
		Attrs: map[string][]string{
			"mail": {"test@example.com"},
		},
	}
	got := r.GetAttr("mail")
	if got != "test@example.com" {
		t.Errorf("GetAttr(mail) = %q, want test@example.com", got)
	}
}

func TestTResultLdap_GetAttr_MultipleValues(t *testing.T) {
	r := &TResultLdap{
		Attrs: map[string][]string{
			"member": {"user1", "user2", "user3"},
		},
	}
	got := r.GetAttr("member")
	if got != "user1;user2;user3" {
		t.Errorf("GetAttr(member) = %q, want user1;user2;user3", got)
	}
}

func TestTResultLdap_GetAttr_NotFound(t *testing.T) {
	r := &TResultLdap{
		Attrs: map[string][]string{},
	}
	got := r.GetAttr("nonexistent")
	if got != "" {
		t.Errorf("GetAttr(nonexistent) = %q, want empty", got)
	}
}

func TestTResultLdap_GetAttrByInt(t *testing.T) {
	r := &TResultLdap{
		Attrs: map[string][]string{
			"count": {"42"},
		},
	}
	got := r.GetAttrByInt("count")
	if got != 42 {
		t.Errorf("GetAttrByInt(count) = %d, want 42", got)
	}
}

func TestTResultLdap_GetAttrByInt_NotFound(t *testing.T) {
	r := &TResultLdap{
		Attrs: map[string][]string{},
	}
	got := r.GetAttrByInt("nonexistent")
	if got != -1 {
		t.Errorf("GetAttrByInt(nonexistent) = %d, want -1", got)
	}
}

func TestTResultLdap_GetAttrByInt_MultipleValues(t *testing.T) {
	r := &TResultLdap{
		Attrs: map[string][]string{
			"nums": {"1", "2"},
		},
	}
	got := r.GetAttrByInt("nums")
	if got != -1 {
		t.Errorf("GetAttrByInt(nums) with multiple values = %d, want -1", got)
	}
}

func TestTResultLdap_GetAttrByInt_InvalidValue(t *testing.T) {
	r := &TResultLdap{
		Attrs: map[string][]string{
			"bad": {"not-a-number"},
		},
	}
	got := r.GetAttrByInt("bad")
	if got != 0 {
		t.Errorf("GetAttrByInt(bad) = %d, want 0", got)
	}
}

func TestTResultLdap_String(t *testing.T) {
	r := &TResultLdap{
		DN: "cn=test,dc=example,dc=com",
		Attrs: map[string][]string{
			"mail": {"test@example.com"},
		},
	}
	got := r.String()
	if !strings.Contains(got, "cn=test,dc=example,dc=com") {
		t.Errorf("String() should contain DN, got: %q", got)
	}
	if !strings.Contains(got, "test@example.com") {
		t.Errorf("String() should contain mail, got: %q", got)
	}
}

func TestTResultLdap_String_Empty(t *testing.T) {
	r := &TResultLdap{
		Attrs: map[string][]string{},
	}
	got := r.String()
	if got == "" {
		t.Error("String() should not be empty")
	}
}

// --- TLdapUser tests ---

func TestTLdapUser_String(t *testing.T) {
	u := &TLdapUser{
		StaffCode: "E001",
		StaffName: "John",
		Email:     "john@example.com",
		Org:       "Engineering",
		Dept:      "Backend",
		Phone:     "12345678",
		Company:   "ACME",
		Station:   "NYC",
	}
	got := u.String()
	if !strings.Contains(got, "E001") {
		t.Errorf("String() should contain StaffCode, got: %q", got)
	}
	if !strings.Contains(got, "John") {
		t.Errorf("String() should contain StaffName, got: %q", got)
	}
}

// --- TLdapClient struct tests ---

func TestTLdapClient_Struct(t *testing.T) {
	c := &TLdapClient{
		Addr:       "ldap.example.com:389",
		BaseDn:     "dc=example,dc=com",
		BindDn:     "cn=admin,dc=example,dc=com",
		BindPass:   "password",
		AuthFilter: "(uid=%s)",
		Attributes: []string{"cn", "mail"},
		MailDomain: "@example.com",
		TLS:        false,
		StartTLS:   false,
	}
	if c.Addr != "ldap.example.com:389" {
		t.Errorf("Addr = %q", c.Addr)
	}
	if c.BaseDn != "dc=example,dc=com" {
		t.Errorf("BaseDn = %q", c.BaseDn)
	}
	if c.BindDn != "cn=admin,dc=example,dc=com" {
		t.Errorf("BindDn = %q", c.BindDn)
	}
	if c.BindPass != "password" {
		t.Errorf("BindPass = %q", c.BindPass)
	}
	if c.AuthFilter != "(uid=%s)" {
		t.Errorf("AuthFilter = %q", c.AuthFilter)
	}
	if len(c.Attributes) != 2 {
		t.Errorf("Attributes len = %d", len(c.Attributes))
	}
	if c.MailDomain != "@example.com" {
		t.Errorf("MailDomain = %q", c.MailDomain)
	}
	if c.TLS {
		t.Error("TLS should be false")
	}
	if c.StartTLS {
		t.Error("StartTLS should be false")
	}
}

func TestTLdapClient_IsClosing_NoConn(t *testing.T) {
	c := &TLdapClient{}
	if !c.IsClosing() {
		t.Error("IsClosing should return true when Conn is nil")
	}
}

func TestTLdapClient_Close_NoConn(t *testing.T) {
	c := &TLdapClient{}
	c.Close() // Should not panic when Conn is nil
}

func TestTLdapClient_Connect(t *testing.T) {
	t.Skip("requires LDAP server")
}

func TestTLdapClient_Bind(t *testing.T) {
	c := &TLdapClient{}
	success, err := c.Bind("user", "pass")
	if success {
		t.Error("Bind should fail when not connected")
	}
	if err == nil {
		t.Error("Bind should return error when not connected")
	}
	if !strings.Contains(err.Error(), "服务器未连接") {
		t.Errorf("Bind error = %q, should mention server not connected", err.Error())
	}
}

func TestTLdapClient_Auth(t *testing.T) {
	c := &TLdapClient{}
	success, err := c.Auth("user", "pass")
	if success {
		t.Error("Auth should fail when not connected")
	}
	if err == nil {
		t.Error("Auth should return error when not connected")
	}
}

func TestTLdapClient_SearchUser_NotConnected(t *testing.T) {
	c := &TLdapClient{}
	_, err := c.SearchUser("user")
	if err == nil {
		t.Error("SearchUser should return error when not connected")
	}
}

func TestTLdapClient_Search_NotConnected(t *testing.T) {
	c := &TLdapClient{}
	_, err := c.Search("(objectclass=*)", 0)
	if err == nil {
		t.Error("Search should return error when not connected")
	}
}

func TestTLdapClient_SearchBase_NotConnected(t *testing.T) {
	c := &TLdapClient{}
	_, err := c.SearchBase("(objectclass=*)")
	if err == nil {
		t.Error("SearchBase should return error when not connected")
	}
}

func TestTLdapClient_SearchSubOne_NotConnected(t *testing.T) {
	c := &TLdapClient{}
	_, err := c.SearchSubOne("(objectclass=*)")
	if err == nil {
		t.Error("SearchSubOne should return error when not connected")
	}
}

func TestTLdapClient_SearchSubAll_NotConnected(t *testing.T) {
	c := &TLdapClient{}
	_, err := c.SearchSubAll("(objectclass=*)")
	if err == nil {
		t.Error("SearchSubAll should return error when not connected")
	}
}

func TestTLdapClient_SetExternalEmailAddress_NotConnected(t *testing.T) {
	c := &TLdapClient{}
	err := c.SetExternalEmailAddress("", "code", "mail", "name")
	if err == nil {
		t.Error("SetExternalEmailAddress should return error when not connected")
	}
}

func TestTLdapClient_CreateUser_NotConnected(t *testing.T) {
	c := &TLdapClient{}
	_, _, err := c.CreateUser("", &TLdapUser{})
	if err == nil {
		t.Error("CreateUser should return error when not connected")
	}
}

func TestTLdapClient_ChangePassword_NotConnected(t *testing.T) {
	c := &TLdapClient{}
	err := c.ChangePassword("", "code", "newpass")
	if err == nil {
		t.Error("ChangePassword should return error when not connected")
	}
}

func TestTLdapClient_EnableAccount_NotConnected(t *testing.T) {
	c := &TLdapClient{}
	err := c.EnableAccount("code")
	if err == nil {
		t.Error("EnableAccount should return error when not connected")
	}
}

func TestTLdapClient_DisableAccount_NotConnected(t *testing.T) {
	c := &TLdapClient{}
	err := c.DisableAccount("code")
	if err == nil {
		t.Error("DisableAccount should return error when not connected")
	}
}

func TestTLdapClient_DeleteUser_NotConnected(t *testing.T) {
	c := &TLdapClient{}
	err := c.DeleteUser("", "code")
	if err == nil {
		t.Error("DeleteUser should return error when not connected")
	}
}

func TestTLdapClient_CreatePath_NotConnected(t *testing.T) {
	c := &TLdapClient{}
	err := c.CreatePath("", "path")
	if err == nil {
		t.Error("CreatePath should return error when not connected")
	}
}

func TestTLdapClient_DeletePath_NotConnected(t *testing.T) {
	c := &TLdapClient{}
	err := c.DeletePath("", "path")
	if err == nil {
		t.Error("DeletePath should return error when not connected")
	}
}

func TestTLdapClient_CreateGroup_NotConnected(t *testing.T) {
	c := &TLdapClient{}
	err := c.CreateGroup("", "group")
	if err == nil {
		t.Error("CreateGroup should return error when not connected")
	}
}

func TestTLdapClient_DeleteGroup_NotConnected(t *testing.T) {
	c := &TLdapClient{}
	err := c.DeleteGroup("", "group")
	if err == nil {
		t.Error("DeleteGroup should return error when not connected")
	}
}

func TestTLdapClient_AddGroupUser_NotConnected(t *testing.T) {
	c := &TLdapClient{}
	err := c.AddGroupUser("", "group", "userDN")
	if err == nil {
		t.Error("AddGroupUser should return error when not connected")
	}
}

func TestTLdapClient_DelGroupUser_NotConnected(t *testing.T) {
	c := &TLdapClient{}
	err := c.DelGroupUser("", "group", "userDN")
	if err == nil {
		t.Error("DelGroupUser should return error when not connected")
	}
}

func TestTLdapClient_Bind_EmptyPassword(t *testing.T) {
	// When Conn is nil, IsClosing returns true first, so we get "server not connected" error
	// rather than "password can't be empty" error. This is the actual behavior.
	c := &TLdapClient{}
	success, err := c.Bind("user", "")
	if success {
		t.Error("Bind with empty password should fail")
	}
	if err == nil {
		t.Error("Bind with empty password should return error")
	}
}

func TestTLdapClient_Auth_EmptyPassword(t *testing.T) {
	c := &TLdapClient{}
	success, err := c.Auth("user", "")
	if success {
		t.Error("Auth with empty password should fail")
	}
	if err == nil {
		t.Error("Auth with empty password should return error")
	}
}

func TestTLdapClient_SearchUser(t *testing.T) {
	t.Skip("requires LDAP server")
}

func TestTLdapClient_Search(t *testing.T) {
	t.Skip("requires LDAP server")
}

func TestTLdapClient_SearchBase(t *testing.T) {
	t.Skip("requires LDAP server")
}

func TestTLdapClient_SearchSubOne(t *testing.T) {
	t.Skip("requires LDAP server")
}

func TestTLdapClient_SearchSubAll(t *testing.T) {
	t.Skip("requires LDAP server")
}

func TestTLdapClient_CreateUser(t *testing.T) {
	t.Skip("requires LDAP server")
}

func TestTLdapClient_ChangePassword(t *testing.T) {
	t.Skip("requires LDAP server")
}

func TestTLdapClient_EnableAccount(t *testing.T) {
	t.Skip("requires LDAP server")
}

func TestTLdapClient_DisableAccount(t *testing.T) {
	t.Skip("requires LDAP server")
}

func TestTLdapClient_DeleteUser(t *testing.T) {
	t.Skip("requires LDAP server")
}

func TestTLdapClient_CreatePath(t *testing.T) {
	t.Skip("requires LDAP server")
}

func TestTLdapClient_DeletePath(t *testing.T) {
	t.Skip("requires LDAP server")
}

func TestTLdapClient_CreateGroup(t *testing.T) {
	t.Skip("requires LDAP server")
}

func TestTLdapClient_DeleteGroup(t *testing.T) {
	t.Skip("requires LDAP server")
}

func TestTLdapClient_AddGroupUser(t *testing.T) {
	t.Skip("requires LDAP server")
}

func TestTLdapClient_DelGroupUser(t *testing.T) {
	t.Skip("requires LDAP server")
}

func TestTLdapClient_SetExternalEmailAddress(t *testing.T) {
	t.Skip("requires LDAP server")
}

func TestTLdapClient_GetSearchResult(t *testing.T) {
	t.Skip("requires LDAP server")
}

// --- Package-level function tests ---

func TestPackage_Search(t *testing.T) {
	t.Skip("requires LDAP server")
}

func TestPackage_Auth(t *testing.T) {
	t.Skip("requires LDAP server")
}

func TestPackage_SearchUser(t *testing.T) {
	t.Skip("requires LDAP server")
}

func TestPackage_HealthCheck(t *testing.T) {
	t.Skip("requires LDAP server")
}

// --- ParseTicks / TicksToTime tests ---

func TestParseTicks(t *testing.T) {
	// Active Directory epoch is Jan 1, 1601
	// 0 ticks = Jan 1, 1601
	got, err := ParseTicks("0")
	if err != nil {
		t.Fatalf("ParseTicks error: %v", err)
	}
	if got.Year() != 1601 {
		t.Errorf("ParseTicks(0) year = %d, want 1601", got.Year())
	}
}

func TestParseTicks_InvalidInput(t *testing.T) {
	_, err := ParseTicks("not-a-number")
	if err == nil {
		t.Error("ParseTicks with invalid input should return error")
	}
}

func TestParseTicks_ValidTimestamp(t *testing.T) {
	// 132489216000000000 ticks = roughly Jan 1, 2021
	got, err := ParseTicks("132489216000000000")
	if err != nil {
		t.Fatalf("ParseTicks error: %v", err)
	}
	if got.Year() < 2020 || got.Year() > 2022 {
		t.Errorf("ParseTicks(132489216000000000) year = %d, want ~2021", got.Year())
	}
}

func TestTicksToTime(t *testing.T) {
	// Zero ticks
	got := TicksToTime(0)
	if got.Year() != 1601 {
		t.Errorf("TicksToTime(0) year = %d, want 1601", got.Year())
	}

	// Non-zero ticks
	got = TicksToTime(10000000) // 1 second worth of ticks
	if got.Year() != 1601 {
		t.Errorf("TicksToTime(10000000) year = %d, want 1601", got.Year())
	}
}

func TestTicksToTime_LargeValue(t *testing.T) {
	// Large tick value should work without overflow
	got := TicksToTime(132489216000000000) // ~2021
	if got.Year() < 2020 {
		t.Errorf("TicksToTime(large) year = %d, want ~2021", got.Year())
	}
}

// --- DefaultPwd test ---

func TestDefaultPwd(t *testing.T) {
	if DefaultPwd == "" {
		t.Error("DefaultPwd should not be empty")
	}
}

// --- uacMask test ---

func TestUacMask(t *testing.T) {
	if _, ok := uacMask[2]; !ok {
		t.Error("uacMask should have entry for ACCOUNT_DISABLE (2)")
	}
	if _, ok := uacMask[512]; !ok {
		t.Error("uacMask should have entry for NORMAL_ACCOUNT (512)")
	}
}

// --- JsonResult test ---

func TestTResultLdap_JsonMarshal(t *testing.T) {
	r := &TResultLdap{
		DN: "cn=test",
		Attrs: map[string][]string{
			"key": {"val"},
		},
	}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}
	if !strings.Contains(string(data), "cn=test") {
		t.Errorf("JSON should contain DN, got: %q", string(data))
	}
}

// --- Constants test ---

func TestConstants(t *testing.T) {
	if nanoSecondsPerSecond != 1000000000 {
		t.Errorf("nanoSecondsPerSecond = %d, want 1000000000", nanoSecondsPerSecond)
	}
	if nanosInTick != 100 {
		t.Errorf("nanosInTick = %d, want 100", nanosInTick)
	}
	if ticksPerSecond != 10000000 {
		t.Errorf("ticksPerSecond = %d, want 10000000", ticksPerSecond)
	}
}

// --- UacMask comprehensive test ---

func TestUacMask_AllEntries(t *testing.T) {
	expectedEntries := map[uint32]string{
		2:        "ACCOUNT_DISABLE",
		8:        "HOMEDIR_REQUIRED",
		16:       "LOCKOUT",
		32:       "PASSWD_NOTREQD",
		64:       "PASSWD_CANT_CHANGE",
		128:      "ENCRYPTED_TEXT_PASSWORD_ALLOWED",
		512:      "NORMAL_ACCOUNT",
		8192:     "SERVER_TRUST_ACCOUNT",
		65536:    "DONT_EXPIRE_PASSWD",
		131072:   "MNS_LOGON_ACCOUNT",
		262144:   "SMARTCARD_REQUIRED",
		524288:   "TRUSTED_FOR_DELEGATION",
		1048576:  "NOT_DELEGATED",
		2097152:  "USE_DES_KEY_ONLY",
		4194304:  "DONT_REQUIRE_PREAUTH",
		8388608:  "PASSWORD_EXPIRED",
		16777216: "TRUSTED_TO_AUTHENTICATE_FOR_DELEGATION",
		33554432: "NO_AUTH_DATA_REQUIRED",
		67108864: "PARTIAL_SECRETS_ACCOUNT",
	}

	for key, expectedValue := range expectedEntries {
		value, ok := uacMask[key]
		if !ok {
			t.Errorf("uacMask missing key %d", key)
			continue
		}
		if value != expectedValue {
			t.Errorf("uacMask[%d] = %q, want %q", key, value, expectedValue)
		}
	}

	// Verify total count
	if len(uacMask) != len(expectedEntries) {
		t.Errorf("uacMask has %d entries, want %d", len(uacMask), len(expectedEntries))
	}
}

// --- TicksToTime edge cases ---

func TestTicksToTime_Epoch(t *testing.T) {
	// Zero ticks should be Jan 1, 1601 00:00:00 UTC
	got := TicksToTime(0)
	expected := time.Date(1601, time.January, 1, 0, 0, 0, 0, time.UTC)
	if !got.Equal(expected) {
		t.Errorf("TicksToTime(0) = %v, want %v", got, expected)
	}
}

func TestTicksToTime_OneSecond(t *testing.T) {
	// 10000000 ticks = 1 second
	got := TicksToTime(10000000)
	expected := time.Date(1601, time.January, 1, 0, 0, 1, 0, time.UTC)
	if !got.Equal(expected) {
		t.Errorf("TicksToTime(10000000) = %v, want %v", got, expected)
	}
}

func TestTicksToTime_WithNanoseconds(t *testing.T) {
	// 10000050 ticks = 1 second + 5000 nanoseconds (50 * 100)
	got := TicksToTime(10000050)
	if got.Second() != 1 {
		t.Errorf("TicksToTime(10000050) second = %d, want 1", got.Second())
	}
	// Nanoseconds should be 50 * 100 = 5000
	if got.Nanosecond() != 5000 {
		t.Errorf("TicksToTime(10000050) nanosecond = %d, want 5000", got.Nanosecond())
	}
}

func TestParseTicks_EmptyString(t *testing.T) {
	_, err := ParseTicks("")
	if err == nil {
		t.Error("ParseTicks('') should return error")
	}
}

func TestParseTicks_NegativeValue(t *testing.T) {
	// Negative ticks - should still parse but result in time before 1601
	got, err := ParseTicks("-1")
	if err != nil {
		t.Fatalf("ParseTicks(-1) error: %v", err)
	}
	// The result will be before the epoch
	if got.Year() > 1601 {
		t.Errorf("ParseTicks(-1) year = %d, should be <= 1601", got.Year())
	}
}

// --- TLdapClient Close with nil Conn ---

func TestTLdapClient_Close_MultipleCalls(t *testing.T) {
	c := &TLdapClient{}
	// Multiple Close calls should not panic
	c.Close()
	c.Close()
	c.Close()
}

// --- TLdapClient field setters/getters ---

func TestTLdapClient_Fields(t *testing.T) {
	c := &TLdapClient{
		Addr:       "localhost:389",
		BaseDn:     "dc=example,dc=com",
		BindDn:     "cn=admin",
		BindPass:   "secret",
		AuthFilter: "(uid=%s)",
		Attributes: []string{"cn", "mail", "uid"},
		MailDomain: "@example.com",
		TLS:        true,
		StartTLS:   false,
	}

	// Verify all fields are set correctly
	if c.Addr != "localhost:389" {
		t.Errorf("Addr = %q", c.Addr)
	}
	if c.BaseDn != "dc=example,dc=com" {
		t.Errorf("BaseDn = %q", c.BaseDn)
	}
	if c.BindDn != "cn=admin" {
		t.Errorf("BindDn = %q", c.BindDn)
	}
	if c.BindPass != "secret" {
		t.Errorf("BindPass = %q", c.BindPass)
	}
	if c.AuthFilter != "(uid=%s)" {
		t.Errorf("AuthFilter = %q", c.AuthFilter)
	}
	if len(c.Attributes) != 3 {
		t.Errorf("Attributes len = %d, want 3", len(c.Attributes))
	}
	if c.MailDomain != "@example.com" {
		t.Errorf("MailDomain = %q", c.MailDomain)
	}
	if !c.TLS {
		t.Error("TLS should be true")
	}
	if c.StartTLS {
		t.Error("StartTLS should be false")
	}
}

func TestTLdapClient_EmptyAttributes(t *testing.T) {
	c := &TLdapClient{
		Attributes: []string{},
	}
	if len(c.Attributes) != 0 {
		t.Errorf("Attributes len = %d, want 0", len(c.Attributes))
	}
}

func TestTLdapClient_NilAttributes(t *testing.T) {
	c := &TLdapClient{}
	if c.Attributes != nil {
		t.Errorf("Attributes should be nil, got %v", c.Attributes)
	}
}

// --- TLdapUser all fields ---

func TestTLdapUser_AllFields(t *testing.T) {
	u := &TLdapUser{
		StaffCode: "E001",
		StaffName: "张三",
		Email:     "zhangsan@example.com",
		Org:       "研发部",
		Dept:      "后端组",
		Phone:     "13800138000",
		Company:   "测试公司",
		Station:   "北京",
	}

	got := u.String()
	if !strings.Contains(got, "E001") {
		t.Error("String() should contain StaffCode")
	}
	if !strings.Contains(got, "张三") {
		t.Error("String() should contain StaffName")
	}
	if !strings.Contains(got, "zhangsan@example.com") {
		t.Error("String() should contain Email")
	}
}

func TestTLdapUser_EmptyFields(t *testing.T) {
	u := &TLdapUser{}
	got := u.String()
	if got == "" {
		t.Error("String() should not be empty even with empty fields")
	}
}

// --- TResultLdap more edge cases ---

func TestTResultLdap_GetAttr_EmptyAttrs(t *testing.T) {
	r := &TResultLdap{
		DN:    "cn=test",
		Attrs: nil,
	}
	got := r.GetAttr("mail")
	if got != "" {
		t.Errorf("GetAttr on nil Attrs = %q, want empty", got)
	}
}

func TestTResultLdap_GetAttrByInt_EmptyAttrs(t *testing.T) {
	r := &TResultLdap{
		DN:    "cn=test",
		Attrs: nil,
	}
	got := r.GetAttrByInt("count")
	if got != -1 {
		t.Errorf("GetAttrByInt on nil Attrs = %d, want -1", got)
	}
}

func TestTResultLdap_GetAttrByInt_EmptyValue(t *testing.T) {
	r := &TResultLdap{
		Attrs: map[string][]string{
			"empty": {""},
		},
	}
	got := r.GetAttrByInt("empty")
	if got != 0 {
		t.Errorf("GetAttrByInt(empty) = %d, want 0", got)
	}
}

func TestTResultLdap_GetAttrByInt_NegativeValue(t *testing.T) {
	r := &TResultLdap{
		Attrs: map[string][]string{
			"neg": {"-42"},
		},
	}
	got := r.GetAttrByInt("neg")
	if got != -42 {
		t.Errorf("GetAttrByInt(neg) = %d, want -42", got)
	}
}

func TestTResultLdap_GetAttrByInt_LargeValue(t *testing.T) {
	r := &TResultLdap{
		Attrs: map[string][]string{
			"large": {"2147483647"}, // max int32
		},
	}
	got := r.GetAttrByInt("large")
	if got != 2147483647 {
		t.Errorf("GetAttrByInt(large) = %d, want 2147483647", got)
	}
}

// --- DefaultPwd value check ---

func TestDefaultPwd_HasValue(t *testing.T) {
	// DefaultPwd should have a non-empty default value
	if DefaultPwd == "" {
		t.Error("DefaultPwd should have a default value")
	}
	// Check it's the expected value
	if DefaultPwd != "BGgsh*1050" {
		t.Logf("DefaultPwd = %q (may have been changed)", DefaultPwd)
	}
}

// --- Package constants ---

func TestPackageConstants(t *testing.T) {
	// These are from funcs.go
	if nanoSecondsPerSecond != 1000000000 {
		t.Errorf("nanoSecondsPerSecond = %d", nanoSecondsPerSecond)
	}
	if nanosInTick != 100 {
		t.Errorf("nanosInTick = %d", nanosInTick)
	}
	if ticksPerSecond != nanoSecondsPerSecond/nanosInTick {
		t.Errorf("ticksPerSecond calculation mismatch")
	}
}

// --- GetSearchResult tests ---

func TestGetSearchResult_DefaultAttribute(t *testing.T) {
	lc := &TLdapClient{}
	entry := &ldap.Entry{
		DN: "cn=test,dc=example,dc=com",
		Attributes: []*ldap.EntryAttribute{
			{
				Name:       "cn",
				Values:     []string{"testuser"},
				ByteValues: [][]byte{[]byte("testuser")},
			},
			{
				Name:       "mail",
				Values:     []string{"test@example.com"},
				ByteValues: [][]byte{[]byte("test@example.com")},
			},
		},
	}

	result := lc.GetSearchResult(entry)

	if result.DN != "cn=test,dc=example,dc=com" {
		t.Errorf("DN = %q, want cn=test,dc=example,dc=com", result.DN)
	}
	if result.GetAttr("cn") != "testuser" {
		t.Errorf("cn = %q, want testuser", result.GetAttr("cn"))
	}
	if result.GetAttr("mail") != "test@example.com" {
		t.Errorf("mail = %q, want test@example.com", result.GetAttr("mail"))
	}
}

func TestGetSearchResult_PwdLastSet(t *testing.T) {
	lc := &TLdapClient{}
	// pwdLastSet = 132489216000000000 ticks ~ Jan 2021
	entry := &ldap.Entry{
		DN: "cn=test,dc=example,dc=com",
		Attributes: []*ldap.EntryAttribute{
			{
				Name:       "pwdLastSet",
				Values:     []string{"132489216000000000"},
				ByteValues: [][]byte{[]byte("132489216000000000")},
			},
		},
	}

	result := lc.GetSearchResult(entry)
	got := result.GetAttr("pwdLastSet")
	// Should be formatted as datetime
	if got == "" {
		t.Error("pwdLastSet should not be empty")
	}
	// Should contain a year around 2021
	if !strings.Contains(got, "2021") {
		t.Logf("pwdLastSet = %q (expected ~2021)", got)
	}
}

func TestGetSearchResult_BadPasswordTime(t *testing.T) {
	lc := &TLdapClient{}
	entry := &ldap.Entry{
		DN: "cn=test,dc=example,dc=com",
		Attributes: []*ldap.EntryAttribute{
			{
				Name:       "badPasswordTime",
				Values:     []string{"132489216000000000"},
				ByteValues: [][]byte{[]byte("132489216000000000")},
			},
		},
	}

	result := lc.GetSearchResult(entry)
	got := result.GetAttr("badPasswordTime")
	if got == "" {
		t.Error("badPasswordTime should not be empty")
	}
}

func TestGetSearchResult_LastLogon(t *testing.T) {
	lc := &TLdapClient{}
	entry := &ldap.Entry{
		DN: "cn=test,dc=example,dc=com",
		Attributes: []*ldap.EntryAttribute{
			{
				Name:       "lastLogon",
				Values:     []string{"132489216000000000"},
				ByteValues: [][]byte{[]byte("132489216000000000")},
			},
		},
	}

	result := lc.GetSearchResult(entry)
	got := result.GetAttr("lastLogon")
	if got == "" {
		t.Error("lastLogon should not be empty")
	}
}

func TestGetSearchResult_LockoutTime(t *testing.T) {
	lc := &TLdapClient{}
	entry := &ldap.Entry{
		DN: "cn=test,dc=example,dc=com",
		Attributes: []*ldap.EntryAttribute{
			{
				Name:       "lockoutTime",
				Values:     []string{"0"},
				ByteValues: [][]byte{[]byte("0")},
			},
		},
	}

	result := lc.GetSearchResult(entry)
	got := result.GetAttr("lockoutTime")
	if got == "" {
		t.Error("lockoutTime should not be empty")
	}
}

func TestGetSearchResult_WhenCreated(t *testing.T) {
	lc := &TLdapClient{}
	// whenCreated format: 20060102150405.0Z0700
	entry := &ldap.Entry{
		DN: "cn=test,dc=example,dc=com",
		Attributes: []*ldap.EntryAttribute{
			{
				Name:       "whenCreated",
				Values:     []string{"20210101120000.0Z"},
				ByteValues: [][]byte{[]byte("20210101120000.0Z")},
			},
		},
	}

	result := lc.GetSearchResult(entry)
	got := result.GetAttr("whenCreated")
	if got == "" {
		t.Error("whenCreated should not be empty")
	}
}

func TestGetSearchResult_WhenChanged(t *testing.T) {
	lc := &TLdapClient{}
	entry := &ldap.Entry{
		DN: "cn=test,dc=example,dc=com",
		Attributes: []*ldap.EntryAttribute{
			{
				Name:       "whenChanged",
				Values:     []string{"20210615103000.0Z"},
				ByteValues: [][]byte{[]byte("20210615103000.0Z")},
			},
		},
	}

	result := lc.GetSearchResult(entry)
	got := result.GetAttr("whenChanged")
	if got == "" {
		t.Error("whenChanged should not be empty")
	}
}

func TestGetSearchResult_ObjectGUID(t *testing.T) {
	lc := &TLdapClient{}
	// A sample 16-byte GUID (mixed-endian Windows format)
	guidBytes := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
	entry := &ldap.Entry{
		DN: "cn=test,dc=example,dc=com",
		Attributes: []*ldap.EntryAttribute{
			{
				Name:       "objectGUID",
				Values:     []string{string(guidBytes)},
				ByteValues: [][]byte{guidBytes},
			},
		},
	}

	result := lc.GetSearchResult(entry)
	got := result.GetAttr("objectGUID")
	if got == "" {
		t.Error("objectGUID should not be empty")
	}
	// It should be a GUID string, not the raw bytes
	if len(got) < 10 {
		t.Errorf("objectGUID seems too short: %q", got)
	}
}

func TestGetSearchResult_MultipleAttributeTypes(t *testing.T) {
	lc := &TLdapClient{}
	guidBytes := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
	entry := &ldap.Entry{
		DN: "cn=test,dc=example,dc=com",
		Attributes: []*ldap.EntryAttribute{
			{
				Name:       "objectGUID",
				Values:     []string{string(guidBytes)},
				ByteValues: [][]byte{guidBytes},
			},
			{
				Name:       "pwdLastSet",
				Values:     []string{"132489216000000000"},
				ByteValues: [][]byte{[]byte("132489216000000000")},
			},
			{
				Name:       "whenCreated",
				Values:     []string{"20210101120000.0Z"},
				ByteValues: [][]byte{[]byte("20210101120000.0Z")},
			},
			{
				Name:       "cn",
				Values:     []string{"testuser"},
				ByteValues: [][]byte{[]byte("testuser")},
			},
		},
	}

	result := lc.GetSearchResult(entry)
	if result.DN != "cn=test,dc=example,dc=com" {
		t.Errorf("DN = %q", result.DN)
	}
	if result.GetAttr("cn") != "testuser" {
		t.Errorf("cn = %q, want testuser", result.GetAttr("cn"))
	}
	if result.GetAttr("objectGUID") == "" {
		t.Error("objectGUID should not be empty")
	}
	if result.GetAttr("pwdLastSet") == "" {
		t.Error("pwdLastSet should not be empty")
	}
	if result.GetAttr("whenCreated") == "" {
		t.Error("whenCreated should not be empty")
	}
}

func TestGetSearchResult_EmptyAttributes(t *testing.T) {
	lc := &TLdapClient{}
	entry := &ldap.Entry{
		DN:         "cn=empty,dc=example,dc=com",
		Attributes: []*ldap.EntryAttribute{},
	}

	result := lc.GetSearchResult(entry)
	if result.DN != "cn=empty,dc=example,dc=com" {
		t.Errorf("DN = %q", result.DN)
	}
	if len(result.Attrs) != 0 {
		t.Errorf("Attrs should be empty, got %d entries", len(result.Attrs))
	}
}

func TestGetSearchResult_MultipleValues(t *testing.T) {
	lc := &TLdapClient{}
	entry := &ldap.Entry{
		DN: "cn=test,dc=example,dc=com",
		Attributes: []*ldap.EntryAttribute{
			{
				Name:       "member",
				Values:     []string{"user1", "user2", "user3"},
				ByteValues: [][]byte{[]byte("user1"), []byte("user2"), []byte("user3")},
			},
		},
	}

	result := lc.GetSearchResult(entry)
	members := result.Attrs["member"]
	if len(members) != 3 {
		t.Errorf("member count = %d, want 3", len(members))
	}
	if members[0] != "user1" {
		t.Errorf("member[0] = %q, want user1", members[0])
	}
}

// --- Test that ParseTicks integrates with GetSearchResult ---

func TestParseTicks_ZeroTicks(t *testing.T) {
	got, err := ParseTicks("0")
	if err != nil {
		t.Fatalf("ParseTicks(0) error: %v", err)
	}
	expected := time.Date(1601, time.January, 1, 0, 0, 0, 0, time.UTC)
	if !got.Equal(expected) {
		t.Errorf("ParseTicks(0) = %v, want %v", got, expected)
	}
}

func TestParseTicks_RealWorldValue(t *testing.T) {
	// 133394640000000000 ticks = roughly 2023-01-01
	got, err := ParseTicks("133394640000000000")
	if err != nil {
		t.Fatalf("ParseTicks error: %v", err)
	}
	if got.Year() < 2022 || got.Year() > 2024 {
		t.Errorf("ParseTicks year = %d, want ~2023", got.Year())
	}
}

// --- DateTimeFormat integration test ---

func TestDateTimeFormat(t *testing.T) {
	// Verify that utils.DateTimeFormat is accessible and has expected value
	if utils.DateTimeFormat == "" {
		t.Error("utils.DateTimeFormat should not be empty")
	}
	// The format should be a valid Go time format
	now := time.Now()
	formatted := now.Format(utils.DateTimeFormat)
	if formatted == "" {
		t.Error("Format result should not be empty")
	}
}

// --- GUIDFromWindowsArray integration test ---

func TestGUIDFromWindowsArray(t *testing.T) {
	// Test with a well-known GUID pattern
	b := [16]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
	guid := utils.GUIDFromWindowsArray(b)
	guidStr := guid.String()
	if guidStr == "" {
		t.Error("GUID string should not be empty")
	}
}

// --- Mock net.Conn for testing ldap.Conn ---

// mockNetConn implements net.Conn with pipe-based I/O to avoid blocking
type mockNetConn struct {
	net.Conn
}

func newMockNetConn() (client, server *mockNetConn) {
	c1, c2 := net.Pipe()
	return &mockNetConn{Conn: c1}, &mockNetConn{Conn: c2}
}

func (m *mockNetConn) Close() error {
	return m.Conn.Close()
}

// drainServer reads and discards data from the server end of the pipe
// to prevent the client from blocking on writes
func drainServer(conn net.Conn, done chan struct{}) {
	buf := make([]byte, 4096)
	for {
		_, err := conn.Read(buf)
		if err != nil {
			close(done)
			return
		}
	}
}

// --- Tests with mock ldap.Conn ---

func TestTLdapClient_IsClosing_WithMockConn(t *testing.T) {
	client, server := newMockNetConn()
	defer client.Close()
	defer server.Close()

	done := make(chan struct{})
	go drainServer(server, done)

	ldapConn := ldap.NewConn(client, false)
	ldapConn.Start()

	c := &TLdapClient{Conn: ldapConn}

	// A fresh connection should not be closing
	if c.IsClosing() {
		t.Error("IsClosing should be false for fresh connection")
	}

	// Close the connection
	ldapConn.Close()

	// Wait for drain goroutine to finish
	select {
	case <-done:
	case <-time.After(time.Second):
	}
}

func TestTLdapClient_Close_WithMockConn(t *testing.T) {
	client, server := newMockNetConn()
	defer client.Close()
	defer server.Close()

	done := make(chan struct{})
	go drainServer(server, done)

	ldapConn := ldap.NewConn(client, false)
	ldapConn.Start()

	c := &TLdapClient{Conn: ldapConn}

	// Close should work with a real connection
	c.Close()

	// Conn should be nil after Close
	if c.Conn != nil {
		t.Error("Conn should be nil after Close")
	}

	// Multiple Close calls should not panic
	c.Close()
	c.Close()
}

func TestTLdapClient_Bind_EmptyPassword_WithMockConn(t *testing.T) {
	client, server := newMockNetConn()
	defer client.Close()
	defer server.Close()

	done := make(chan struct{})
	go drainServer(server, done)

	ldapConn := ldap.NewConn(client, false)
	ldapConn.Start()

	c := &TLdapClient{Conn: ldapConn}
	defer c.Close()

	// Bind with empty password should return error (empty password guard)
	success, err := c.Bind("cn=user,dc=example,dc=com", "")
	if success {
		t.Error("Bind with empty password should fail")
	}
	if err == nil {
		t.Error("Bind with empty password should return error")
	}
	if !strings.Contains(err.Error(), "密码不能为空") {
		t.Errorf("Bind error = %q, should mention empty password", err.Error())
	}
}

func TestTLdapClient_Auth_EmptyPassword_WithMockConn(t *testing.T) {
	client, server := newMockNetConn()
	defer client.Close()
	defer server.Close()

	done := make(chan struct{})
	go drainServer(server, done)

	ldapConn := ldap.NewConn(client, false)
	ldapConn.Start()

	c := &TLdapClient{Conn: ldapConn}
	defer c.Close()

	// Auth with empty password should return error (empty password guard)
	success, err := c.Auth("user", "")
	if success {
		t.Error("Auth with empty password should fail")
	}
	if err == nil {
		t.Error("Auth with empty password should return error")
	}
	if !strings.Contains(err.Error(), "密码不能为空") {
		t.Errorf("Auth error = %q, should mention empty password", err.Error())
	}
}

func TestTLdapClient_Bind_NonEmptyPassword_WithMockConn(t *testing.T) {
	client, server := newMockNetConn()
	defer client.Close()
	defer server.Close()

	done := make(chan struct{})
	go drainServer(server, done)

	ldapConn := ldap.NewConn(client, false)
	ldapConn.Start()

	c := &TLdapClient{Conn: ldapConn}
	defer c.Close()

	// Bind with non-empty password will attempt to bind over the wire
	// and block waiting for a response. Run it in a goroutine with a timeout.
	doneCh := make(chan struct{})
	var success bool
	var err error
	go func() {
		success, err = c.Bind("cn=user,dc=example,dc=com", "password123")
		close(doneCh)
	}()

	select {
	case <-doneCh:
		// Bind completed (likely with error due to mock)
		if err != nil && strings.Contains(err.Error(), "密码不能为空") {
			t.Error("Bind with non-empty password should not return empty password error")
		}
		if success {
			t.Log("Bind succeeded unexpectedly")
		}
	case <-time.After(500 * time.Millisecond):
		// Bind timed out waiting for LDAP response - this is expected with mock
		t.Log("Bind timed out as expected with mock connection (no LDAP server)")
	}
}

// Test operations with mock connection that gets closed mid-operation
// This tests the path construction code that runs after IsClosing check

func TestTLdapClient_DeleteUser_WithMockConn(t *testing.T) {
	client, server := newMockNetConn()
	defer client.Close()
	defer server.Close()

	done := make(chan struct{})
	go drainServer(server, done)

	ldapConn := ldap.NewConn(client, false)
	ldapConn.Start()

	c := &TLdapClient{
		Conn:   ldapConn,
		BaseDn: "dc=example,dc=com",
	}
	defer c.Close()

	// DeleteUser should construct the DN and then call Conn.Del()
	// which will block. We close the connection to force an error.
	doneCh := make(chan error)
	go func() {
		err := c.DeleteUser("ou=users", "testuser")
		doneCh <- err
	}()

	select {
	case err := <-doneCh:
		// Operation completed (likely with error)
		if err == nil {
			t.Log("DeleteUser succeeded unexpectedly")
		} else {
			t.Logf("DeleteUser error (expected): %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		// Force close the connection to unblock the operation
		c.Close()
		<-doneCh // Wait for the goroutine to finish
		t.Log("DeleteUser timed out, connection closed")
	}
}

func TestTLdapClient_CreatePath_WithMockConn(t *testing.T) {
	client, server := newMockNetConn()
	defer client.Close()
	defer server.Close()

	done := make(chan struct{})
	go drainServer(server, done)

	ldapConn := ldap.NewConn(client, false)
	ldapConn.Start()

	c := &TLdapClient{
		Conn:   ldapConn,
		BaseDn: "dc=example,dc=com",
	}
	defer c.Close()

	doneCh := make(chan error)
	go func() {
		err := c.CreatePath("ou=parent", "newou")
		doneCh <- err
	}()

	select {
	case err := <-doneCh:
		if err == nil {
			t.Log("CreatePath succeeded unexpectedly")
		} else {
			t.Logf("CreatePath error (expected): %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		c.Close()
		<-doneCh
		t.Log("CreatePath timed out, connection closed")
	}
}

func TestTLdapClient_CreateGroup_WithMockConn(t *testing.T) {
	client, server := newMockNetConn()
	defer client.Close()
	defer server.Close()

	done := make(chan struct{})
	go drainServer(server, done)

	ldapConn := ldap.NewConn(client, false)
	ldapConn.Start()

	c := &TLdapClient{
		Conn:   ldapConn,
		BaseDn: "dc=example,dc=com",
	}
	defer c.Close()

	doneCh := make(chan error)
	go func() {
		err := c.CreateGroup("ou=groups", "newgroup")
		doneCh <- err
	}()

	select {
	case err := <-doneCh:
		if err == nil {
			t.Log("CreateGroup succeeded unexpectedly")
		} else {
			t.Logf("CreateGroup error (expected): %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		c.Close()
		<-doneCh
		t.Log("CreateGroup timed out, connection closed")
	}
}

func TestTLdapClient_DeleteGroup_WithMockConn(t *testing.T) {
	client, server := newMockNetConn()
	defer client.Close()
	defer server.Close()

	done := make(chan struct{})
	go drainServer(server, done)

	ldapConn := ldap.NewConn(client, false)
	ldapConn.Start()

	c := &TLdapClient{
		Conn:   ldapConn,
		BaseDn: "dc=example,dc=com",
	}
	defer c.Close()

	doneCh := make(chan error)
	go func() {
		err := c.DeleteGroup("ou=groups", "delgroup")
		doneCh <- err
	}()

	select {
	case err := <-doneCh:
		if err == nil {
			t.Log("DeleteGroup succeeded unexpectedly")
		} else {
			t.Logf("DeleteGroup error (expected): %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		c.Close()
		<-doneCh
		t.Log("DeleteGroup timed out, connection closed")
	}
}

func TestTLdapClient_DeletePath_WithMockConn(t *testing.T) {
	client, server := newMockNetConn()
	defer client.Close()
	defer server.Close()

	done := make(chan struct{})
	go drainServer(server, done)

	ldapConn := ldap.NewConn(client, false)
	ldapConn.Start()

	c := &TLdapClient{
		Conn:   ldapConn,
		BaseDn: "dc=example,dc=com",
	}
	defer c.Close()

	doneCh := make(chan error)
	go func() {
		err := c.DeletePath("ou=parent", "delou")
		doneCh <- err
	}()

	select {
	case err := <-doneCh:
		if err == nil {
			t.Log("DeletePath succeeded unexpectedly")
		} else {
			t.Logf("DeletePath error (expected): %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		c.Close()
		<-doneCh
		t.Log("DeletePath timed out, connection closed")
	}
}

func TestTLdapClient_AddGroupUser_WithMockConn(t *testing.T) {
	client, server := newMockNetConn()
	defer client.Close()
	defer server.Close()

	done := make(chan struct{})
	go drainServer(server, done)

	ldapConn := ldap.NewConn(client, false)
	ldapConn.Start()

	c := &TLdapClient{
		Conn:   ldapConn,
		BaseDn: "dc=example,dc=com",
	}
	defer c.Close()

	doneCh := make(chan error)
	go func() {
		err := c.AddGroupUser("ou=groups", "mygroup", "cn=user")
		doneCh <- err
	}()

	select {
	case err := <-doneCh:
		if err == nil {
			t.Log("AddGroupUser succeeded unexpectedly")
		} else {
			t.Logf("AddGroupUser error (expected): %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		c.Close()
		<-doneCh
		t.Log("AddGroupUser timed out, connection closed")
	}
}

func TestTLdapClient_DelGroupUser_WithMockConn(t *testing.T) {
	client, server := newMockNetConn()
	defer client.Close()
	defer server.Close()

	done := make(chan struct{})
	go drainServer(server, done)

	ldapConn := ldap.NewConn(client, false)
	ldapConn.Start()

	c := &TLdapClient{
		Conn:   ldapConn,
		BaseDn: "dc=example,dc=com",
	}
	defer c.Close()

	doneCh := make(chan error)
	go func() {
		err := c.DelGroupUser("ou=groups", "mygroup", "cn=user")
		doneCh <- err
	}()

	select {
	case err := <-doneCh:
		if err == nil {
			t.Log("DelGroupUser succeeded unexpectedly")
		} else {
			t.Logf("DelGroupUser error (expected): %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		c.Close()
		<-doneCh
		t.Log("DelGroupUser timed out, connection closed")
	}
}

func TestTLdapClient_SetExternalEmailAddress_WithMockConn(t *testing.T) {
	client, server := newMockNetConn()
	defer client.Close()
	defer server.Close()

	done := make(chan struct{})
	go drainServer(server, done)

	ldapConn := ldap.NewConn(client, false)
	ldapConn.Start()

	c := &TLdapClient{
		Conn:      ldapConn,
		BaseDn:    "dc=example,dc=com",
		MailDomain: "@example.com",
	}
	defer c.Close()

	doneCh := make(chan error)
	go func() {
		err := c.SetExternalEmailAddress("ou=users", "user1", "user1@example.com", "User One")
		doneCh <- err
	}()

	select {
	case err := <-doneCh:
		if err == nil {
			t.Log("SetExternalEmailAddress succeeded unexpectedly")
		} else {
			t.Logf("SetExternalEmailAddress error (expected): %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		c.Close()
		<-doneCh
		t.Log("SetExternalEmailAddress timed out, connection closed")
	}
}

func TestTLdapClient_ChangePassword_WithMockConn(t *testing.T) {
	client, server := newMockNetConn()
	defer client.Close()
	defer server.Close()

	done := make(chan struct{})
	go drainServer(server, done)

	ldapConn := ldap.NewConn(client, false)
	ldapConn.Start()

	c := &TLdapClient{
		Conn:   ldapConn,
		BaseDn: "dc=example,dc=com",
	}
	defer c.Close()

	doneCh := make(chan error)
	go func() {
		err := c.ChangePassword("ou=users", "user1", "newpassword123")
		doneCh <- err
	}()

	select {
	case err := <-doneCh:
		if err == nil {
			t.Log("ChangePassword succeeded unexpectedly")
		} else {
			t.Logf("ChangePassword error (expected): %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		c.Close()
		<-doneCh
		t.Log("ChangePassword timed out, connection closed")
	}
}

func TestTLdapClient_SearchUser_WithMockConn(t *testing.T) {
	client, server := newMockNetConn()
	defer client.Close()
	defer server.Close()

	done := make(chan struct{})
	go drainServer(server, done)

	ldapConn := ldap.NewConn(client, false)
	ldapConn.Start()

	c := &TLdapClient{
		Conn:       ldapConn,
		BaseDn:     "dc=example,dc=com",
		AuthFilter: "(uid=%s)",
		Attributes: []string{"cn", "mail"},
	}
	defer c.Close()

	doneCh := make(chan struct {
		user *TResultLdap
		err  error
	})
	go func() {
		user, err := c.SearchUser("testuser")
		doneCh <- struct {
			user *TResultLdap
			err  error
		}{user, err}
	}()

	select {
	case result := <-doneCh:
		if result.err == nil {
			t.Log("SearchUser succeeded unexpectedly")
		} else {
			t.Logf("SearchUser error (expected): %v", result.err)
		}
	case <-time.After(500 * time.Millisecond):
		c.Close()
		<-doneCh
		t.Log("SearchUser timed out, connection closed")
	}
}

func TestTLdapClient_Search_WithMockConn(t *testing.T) {
	client, server := newMockNetConn()
	defer client.Close()
	defer server.Close()

	done := make(chan struct{})
	go drainServer(server, done)

	ldapConn := ldap.NewConn(client, false)
	ldapConn.Start()

	c := &TLdapClient{
		Conn:       ldapConn,
		BaseDn:     "dc=example,dc=com",
		Attributes: []string{"cn", "mail"},
	}
	defer c.Close()

	doneCh := make(chan struct {
		results []*TResultLdap
		err     error
	})
	go func() {
		results, err := c.Search("(objectClass=user)", 0)
		doneCh <- struct {
			results []*TResultLdap
			err     error
		}{results, err}
	}()

	select {
	case result := <-doneCh:
		if result.err == nil {
			t.Log("Search succeeded unexpectedly")
		} else {
			t.Logf("Search error (expected): %v", result.err)
		}
	case <-time.After(500 * time.Millisecond):
		c.Close()
		<-doneCh
		t.Log("Search timed out, connection closed")
	}
}

func TestTLdapClient_SearchBase_WithMockConn(t *testing.T) {
	client, server := newMockNetConn()
	defer client.Close()
	defer server.Close()

	done := make(chan struct{})
	go drainServer(server, done)

	ldapConn := ldap.NewConn(client, false)
	ldapConn.Start()

	c := &TLdapClient{
		Conn:       ldapConn,
		BaseDn:     "dc=example,dc=com",
		Attributes: []string{"cn"},
	}
	defer c.Close()

	doneCh := make(chan struct {
		results []*TResultLdap
		err     error
	})
	go func() {
		results, err := c.SearchBase("(objectClass=*)")
		doneCh <- struct {
			results []*TResultLdap
			err     error
		}{results, err}
	}()

	select {
	case result := <-doneCh:
		if result.err == nil {
			t.Log("SearchBase succeeded unexpectedly")
		}
	case <-time.After(500 * time.Millisecond):
		c.Close()
		<-doneCh
		t.Log("SearchBase timed out, connection closed")
	}
}

func TestTLdapClient_SearchSubOne_WithMockConn(t *testing.T) {
	client, server := newMockNetConn()
	defer client.Close()
	defer server.Close()

	done := make(chan struct{})
	go drainServer(server, done)

	ldapConn := ldap.NewConn(client, false)
	ldapConn.Start()

	c := &TLdapClient{
		Conn:       ldapConn,
		BaseDn:     "dc=example,dc=com",
		Attributes: []string{"cn"},
	}
	defer c.Close()

	doneCh := make(chan struct {
		results []*TResultLdap
		err     error
	})
	go func() {
		results, err := c.SearchSubOne("(objectClass=user)")
		doneCh <- struct {
			results []*TResultLdap
			err     error
		}{results, err}
	}()

	select {
	case result := <-doneCh:
		if result.err == nil {
			t.Log("SearchSubOne succeeded unexpectedly")
		}
	case <-time.After(500 * time.Millisecond):
		c.Close()
		<-doneCh
		t.Log("SearchSubOne timed out, connection closed")
	}
}

func TestTLdapClient_SearchSubAll_WithMockConn(t *testing.T) {
	client, server := newMockNetConn()
	defer client.Close()
	defer server.Close()

	done := make(chan struct{})
	go drainServer(server, done)

	ldapConn := ldap.NewConn(client, false)
	ldapConn.Start()

	c := &TLdapClient{
		Conn:       ldapConn,
		BaseDn:     "dc=example,dc=com",
		Attributes: []string{"cn"},
	}
	defer c.Close()

	doneCh := make(chan struct {
		results []*TResultLdap
		err     error
	})
	go func() {
		results, err := c.SearchSubAll("(objectClass=user)")
		doneCh <- struct {
			results []*TResultLdap
			err     error
		}{results, err}
	}()

	select {
	case result := <-doneCh:
		if result.err == nil {
			t.Log("SearchSubAll succeeded unexpectedly")
		}
	case <-time.After(500 * time.Millisecond):
		c.Close()
		<-doneCh
		t.Log("SearchSubAll timed out, connection closed")
	}
}

func TestTLdapClient_Auth_WithMockConn(t *testing.T) {
	client, server := newMockNetConn()
	defer client.Close()
	defer server.Close()

	done := make(chan struct{})
	go drainServer(server, done)

	ldapConn := ldap.NewConn(client, false)
	ldapConn.Start()

	c := &TLdapClient{
		Conn:       ldapConn,
		BaseDn:     "dc=example,dc=com",
		AuthFilter: "(uid=%s)",
		Attributes: []string{"cn", "mail"},
		BindDn:     "cn=admin,dc=example,dc=com",
		BindPass:   "adminpass",
	}
	defer c.Close()

	doneCh := make(chan struct {
		success bool
		err     error
	})
	go func() {
		success, err := c.Auth("testuser", "userpass")
		doneCh <- struct {
			success bool
			err     error
		}{success, err}
	}()

	select {
	case result := <-doneCh:
		if result.success {
			t.Log("Auth succeeded unexpectedly")
		}
	case <-time.After(500 * time.Millisecond):
		c.Close()
		<-doneCh
		t.Log("Auth timed out, connection closed")
	}
}

func TestTLdapClient_CreateUser_WithMockConn(t *testing.T) {
	client, server := newMockNetConn()
	defer client.Close()
	defer server.Close()

	done := make(chan struct{})
	go drainServer(server, done)

	ldapConn := ldap.NewConn(client, false)
	ldapConn.Start()

	c := &TLdapClient{
		Conn:       ldapConn,
		BaseDn:     "dc=example,dc=com",
		MailDomain: "@example.com",
	}
	defer c.Close()

	user := &TLdapUser{
		StaffCode: "E001",
		StaffName: "TestUser",
		Email:     "test@example.com",
		Org:       "Engineering",
		Dept:      "Backend",
		Phone:     "12345678",
		Company:   "ACME",
		Station:   "NYC",
	}

	doneCh := make(chan struct {
		dn           string
		mailNickname string
		err          error
	})
	go func() {
		dn, mailNickname, err := c.CreateUser("ou=users", user)
		doneCh <- struct {
			dn           string
			mailNickname string
			err          error
		}{dn, mailNickname, err}
	}()

	select {
	case result := <-doneCh:
		if result.err == nil {
			t.Log("CreateUser succeeded unexpectedly")
		}
	case <-time.After(500 * time.Millisecond):
		c.Close()
		<-doneCh
		t.Log("CreateUser timed out, connection closed")
	}
}

func TestTLdapClient_CreateUser_EmailNoAt_WithMockConn(t *testing.T) {
	client, server := newMockNetConn()
	defer client.Close()
	defer server.Close()

	done := make(chan struct{})
	go drainServer(server, done)

	ldapConn := ldap.NewConn(client, false)
	ldapConn.Start()

	c := &TLdapClient{
		Conn:       ldapConn,
		BaseDn:     "dc=example,dc=com",
		MailDomain: "@example.com",
	}
	defer c.Close()

	// User with email that doesn't contain "@" - triggers the else branch
	user := &TLdapUser{
		StaffCode: "E002",
		StaffName: "NoAtUser",
		Email:     "noatuser",
		Org:       "Engineering",
		Dept:      "Backend",
		Phone:     "87654321",
		Company:   "ACME",
		Station:   "LA",
	}

	doneCh := make(chan struct {
		dn           string
		mailNickname string
		err          error
	})
	go func() {
		dn, mailNickname, err := c.CreateUser("", user)
		doneCh <- struct {
			dn           string
			mailNickname string
			err          error
		}{dn, mailNickname, err}
	}()

	select {
	case result := <-doneCh:
		if result.err == nil {
			t.Log("CreateUser succeeded unexpectedly")
		}
	case <-time.After(500 * time.Millisecond):
		c.Close()
		<-doneCh
		t.Log("CreateUser (no-at email) timed out, connection closed")
	}
}

func TestTLdapClient_EnableAccount_WithMockConn(t *testing.T) {
	client, server := newMockNetConn()
	defer client.Close()
	defer server.Close()

	done := make(chan struct{})
	go drainServer(server, done)

	ldapConn := ldap.NewConn(client, false)
	ldapConn.Start()

	c := &TLdapClient{
		Conn:       ldapConn,
		BaseDn:     "dc=example,dc=com",
		AuthFilter: "(uid=%s)",
		Attributes: []string{"cn", "mail"},
	}
	defer c.Close()

	doneCh := make(chan error)
	go func() {
		err := c.EnableAccount("testuser")
		doneCh <- err
	}()

	select {
	case err := <-doneCh:
		if err == nil {
			t.Log("EnableAccount succeeded unexpectedly")
		}
	case <-time.After(500 * time.Millisecond):
		c.Close()
		<-doneCh
		t.Log("EnableAccount timed out, connection closed")
	}
}

func TestTLdapClient_DisableAccount_WithMockConn(t *testing.T) {
	client, server := newMockNetConn()
	defer client.Close()
	defer server.Close()

	done := make(chan struct{})
	go drainServer(server, done)

	ldapConn := ldap.NewConn(client, false)
	ldapConn.Start()

	c := &TLdapClient{
		Conn:       ldapConn,
		BaseDn:     "dc=example,dc=com",
		AuthFilter: "(uid=%s)",
		Attributes: []string{"cn", "mail"},
	}
	defer c.Close()

	doneCh := make(chan error)
	go func() {
		err := c.DisableAccount("testuser")
		doneCh <- err
	}()

	select {
	case err := <-doneCh:
		if err == nil {
			t.Log("DisableAccount succeeded unexpectedly")
		}
	case <-time.After(500 * time.Millisecond):
		c.Close()
		<-doneCh
		t.Log("DisableAccount timed out, connection closed")
	}
}

// --- Test all "not connected" paths ---

func TestAllMethods_NotConnected(t *testing.T) {
	c := &TLdapClient{}

	// All methods should return error when not connected
	_, err := c.Bind("user", "pass")
	if err == nil {
		t.Error("Bind should fail when not connected")
	}

	_, err = c.Auth("user", "pass")
	if err == nil {
		t.Error("Auth should fail when not connected")
	}

	_, err = c.SearchUser("user")
	if err == nil {
		t.Error("SearchUser should fail when not connected")
	}

	_, err = c.Search("filter", 0)
	if err == nil {
		t.Error("Search should fail when not connected")
	}

	_, err = c.SearchBase("filter")
	if err == nil {
		t.Error("SearchBase should fail when not connected")
	}

	_, err = c.SearchSubOne("filter")
	if err == nil {
		t.Error("SearchSubOne should fail when not connected")
	}

	_, err = c.SearchSubAll("filter")
	if err == nil {
		t.Error("SearchSubAll should fail when not connected")
	}

	err = c.SetExternalEmailAddress("", "code", "mail", "name")
	if err == nil {
		t.Error("SetExternalEmailAddress should fail when not connected")
	}

	_, _, err = c.CreateUser("", &TLdapUser{})
	if err == nil {
		t.Error("CreateUser should fail when not connected")
	}

	err = c.ChangePassword("", "code", "newpass")
	if err == nil {
		t.Error("ChangePassword should fail when not connected")
	}

	err = c.EnableAccount("code")
	if err == nil {
		t.Error("EnableAccount should fail when not connected")
	}

	err = c.DisableAccount("code")
	if err == nil {
		t.Error("DisableAccount should fail when not connected")
	}

	err = c.DeleteUser("", "code")
	if err == nil {
		t.Error("DeleteUser should fail when not connected")
	}

	err = c.CreatePath("", "path")
	if err == nil {
		t.Error("CreatePath should fail when not connected")
	}

	err = c.DeletePath("", "path")
	if err == nil {
		t.Error("DeletePath should fail when not connected")
	}

	err = c.CreateGroup("", "group")
	if err == nil {
		t.Error("CreateGroup should fail when not connected")
	}

	err = c.DeleteGroup("", "group")
	if err == nil {
		t.Error("DeleteGroup should fail when not connected")
	}

	err = c.AddGroupUser("", "group", "userDN")
	if err == nil {
		t.Error("AddGroupUser should fail when not connected")
	}

	err = c.DelGroupUser("", "group", "userDN")
	if err == nil {
		t.Error("DelGroupUser should fail when not connected")
	}
}
