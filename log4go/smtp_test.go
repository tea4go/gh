package logs

import (
	"encoding/json"
	"testing"
)

func TestSMTPNewWriter(t *testing.T) {
	sw := newSMTPWriter().(*SMTPWriter)
	if sw == nil {
		t.Fatal("newSMTPWriter returned nil")
	}
	if sw.Level != LevelNotice {
		t.Errorf("Level = %d, want %d", sw.Level, LevelNotice)
	}
}

func TestSMTPInit(t *testing.T) {
	sw := newSMTPWriter().(*SMTPWriter)
	config := `{"username":"test@example.com","password":"pass123","host":"smtp.example.com:587","subject":"test subject","fromAddress":"from@example.com","sendTos":["to1@example.com","to2@example.com"],"level":3}`
	err := sw.Init(config)
	if err != nil {
		t.Fatal(err)
	}
	if sw.Username != "test@example.com" {
		t.Errorf("Username = %s, want test@example.com", sw.Username)
	}
	if sw.Host != "smtp.example.com:587" {
		t.Errorf("Host = %s, want smtp.example.com:587", sw.Host)
	}
	if sw.Subject != "test subject" {
		t.Errorf("Subject = %s, want test subject", sw.Subject)
	}
	if len(sw.RecipientAddresses) != 2 {
		t.Errorf("len(RecipientAddresses) = %d, want 2", len(sw.RecipientAddresses))
	}
	if sw.Level != 3 {
		t.Errorf("Level = %d, want 3", sw.Level)
	}
}

func TestSMTPInitBadJSON(t *testing.T) {
	sw := newSMTPWriter().(*SMTPWriter)
	err := sw.Init("{bad json")
	if err == nil {
		t.Fatal("expected error for bad JSON")
	}
}

func TestSMTPGetSMTPAuthWithCredentials(t *testing.T) {
	sw := &SMTPWriter{
		Username: "user@example.com",
		Password: "password",
	}
	auth := sw.getSMTPAuth("smtp.example.com")
	if auth == nil {
		t.Fatal("getSMTPAuth should return auth when credentials are provided")
	}
}

func TestSMTPGetSMTPAuthNoCredentials(t *testing.T) {
	sw := &SMTPWriter{
		Username: "  ",
		Password: "  ",
	}
	auth := sw.getSMTPAuth("smtp.example.com")
	if auth != nil {
		t.Fatal("getSMTPAuth should return nil when no credentials")
	}
}

func TestSMTPWriteMsgFiltered(t *testing.T) {
	sw := &SMTPWriter{
		Level: LevelError,
	}
	err := sw.WriteMsg("test.go", 10, 4, "TestFunc", LevelInfo, GetNow(), "filtered message")
	if err != nil {
		t.Fatal(err)
	}
	// Info > Error, so message should be filtered (returns nil)
}

func TestSMTPWriteMsgSendFail(t *testing.T) {
	sw := &SMTPWriter{
		Username:           "test@example.com",
		Password:           "password",
		Host:               "smtp.nonexistent.invalid:587",
		Subject:            "test",
		FromAddress:        "from@example.com",
		RecipientAddresses: []string{"to@example.com"},
		Level:              LevelEmergency,
	}
	// This will fail to connect, but should not panic
	err := sw.WriteMsg("test.go", 10, 4, "TestFunc", LevelEmergency, GetNow(), "test message")
	if err == nil {
		t.Log("WriteMsg succeeded (unexpected but ok)")
	} else {
		t.Logf("WriteMsg failed as expected: %v", err)
	}
}

func TestSMTPFlush(t *testing.T) {
	sw := newSMTPWriter().(*SMTPWriter)
	sw.Flush() // Should not panic
}

func TestSMTPDestroy(t *testing.T) {
	sw := newSMTPWriter().(*SMTPWriter)
	sw.Destroy() // Should not panic
}

func TestSMTPSetGetLevel(t *testing.T) {
	sw := newSMTPWriter().(*SMTPWriter)
	sw.SetLevel(LevelError)
	if sw.GetLevel() != LevelError {
		t.Errorf("GetLevel = %d, want %d", sw.GetLevel(), LevelError)
	}
}

func TestSMTPConfigParsing(t *testing.T) {
	config := SMTPWriter{}
	jsonStr := `{"username":"u","password":"p","host":"h:25","subject":"s","fromAddress":"f","sendTos":["r1"],"level":5}`
	err := json.Unmarshal([]byte(jsonStr), &config)
	if err != nil {
		t.Fatal(err)
	}
	if config.Username != "u" {
		t.Errorf("Username = %s, want u", config.Username)
	}
	if config.FromAddress != "f" {
		t.Errorf("FromAddress = %s, want f", config.FromAddress)
	}
}
