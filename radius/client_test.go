package radius

import (
	"testing"
	"time"
)

// TestClientDefaults tests Client default values
func TestClientDefaults(t *testing.T) {
	client := &Client{}

	if client.Net != "" {
		t.Errorf("Expected empty Net, got '%s'", client.Net)
	}
	if client.LocalAddr != nil {
		t.Error("Expected nil LocalAddr")
	}
	if client.DialTimeout != 0 {
		t.Errorf("Expected 0 DialTimeout, got %v", client.DialTimeout)
	}
	if client.ReadTimeout != 0 {
		t.Errorf("Expected 0 ReadTimeout, got %v", client.ReadTimeout)
	}
	if client.WriteTimeout != 0 {
		t.Errorf("Expected 0 WriteTimeout, got %v", client.WriteTimeout)
	}
}

// TestClientWithTimeouts tests Client with custom timeouts
func TestClientWithTimeouts(t *testing.T) {
	client := &Client{
		Net:          "udp",
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 2 * time.Second,
	}

	if client.Net != "udp" {
		t.Errorf("Expected 'udp', got '%s'", client.Net)
	}
	if client.DialTimeout != 5*time.Second {
		t.Errorf("Expected 5s DialTimeout, got %v", client.DialTimeout)
	}
	if client.ReadTimeout != 3*time.Second {
		t.Errorf("Expected 3s ReadTimeout, got %v", client.ReadTimeout)
	}
	if client.WriteTimeout != 2*time.Second {
		t.Errorf("Expected 2s WriteTimeout, got %v", client.WriteTimeout)
	}
}

// TestClientSendPacketInvalidAddress tests SendPacket with invalid address
func TestClientSendPacketInvalidAddress(t *testing.T) {
	client := &Client{
		Net:         "udp",
		DialTimeout: 1 * time.Second,
	}

	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	// This should fail with connection error
	_, err := client.SendPacket(packet, "invalid:address:format")
	if err == nil {
		t.Error("Expected error for invalid address")
	}
}
