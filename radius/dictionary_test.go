package radius

import (
	"testing"
)

// TestBuiltinDictionary tests that Builtin dictionary is initialized
func TestBuiltinDictionary(t *testing.T) {
	if Builtin == nil {
		t.Fatal("Builtin dictionary should not be nil")
	}
	if Builtin.NameItems == nil {
		t.Error("Builtin NameItems should not be nil")
	}
}

// TestDictionaryGetName tests GetName method
func TestDictionaryGetName(t *testing.T) {
	name, ok := Builtin.GetName(1)
	if !ok {
		t.Error("GetName should find User-Name")
	}
	if name != "User-Name" {
		t.Errorf("Expected 'User-Name', got '%s'", name)
	}

	// Unknown type
	name, ok = Builtin.GetName(250)
	if ok {
		t.Error("GetName should return false for unknown type")
	}
	if name != "" {
		t.Errorf("Expected empty name for unknown type, got '%s'", name)
	}
}

// TestDictionaryGetIndex tests GetIndex method
func TestDictionaryGetIndex(t *testing.T) {
	id, ok := Builtin.GetIndex("User-Name")
	if !ok {
		t.Error("GetIndex should find User-Name")
	}
	if id != 1 {
		t.Errorf("Expected ID 1, got %d", id)
	}

	// Unknown name
	id, ok = Builtin.GetIndex("UnknownAttribute")
	if ok {
		t.Error("GetIndex should return false for unknown name")
	}
	if id != 0 {
		t.Errorf("Expected ID 0 for unknown name, got %d", id)
	}
}

// TestDictionaryGetFunc tests GetFunc method
func TestDictionaryGetFunc(t *testing.T) {
	codec := Builtin.GetFunc(1) // User-Name
	if codec == nil {
		t.Error("GetFunc should return codec for User-Name")
	}
	if codec == AttributeUnknown {
		t.Error("GetFunc should not return AttributeUnknown for User-Name")
	}

	// Unknown type should return AttributeUnknown
	codec = Builtin.GetFunc(250)
	if codec != AttributeUnknown {
		t.Error("GetFunc should return AttributeUnknown for unknown type")
	}
}

// TestDictionaryNewAttr tests NewAttr method
func TestDictionaryNewAttr(t *testing.T) {
	attr, err := Builtin.NewAttr("User-Name", "testuser")
	if err != nil {
		t.Fatalf("NewAttr failed: %v", err)
	}
	if attr.AttrId != 1 {
		t.Errorf("Expected AttrId 1, got %d", attr.AttrId)
	}
	if attr.AttrValue != "testuser" {
		t.Errorf("Expected 'testuser', got %v", attr.AttrValue)
	}
}

// TestDictionaryNewAttrUnknown tests NewAttr with unknown attribute name
func TestDictionaryNewAttrUnknown(t *testing.T) {
	_, err := Builtin.NewAttr("UnknownAttribute", "value")
	if err == nil {
		t.Error("Expected error for unknown attribute name")
	}
}

// TestDictionaryNewAttrWithTransformer tests NewAttr with transformer attribute
func TestDictionaryNewAttrWithTransformer(t *testing.T) {
	// User-Password has a transformer
	attr, err := Builtin.NewAttr("User-Password", "testpass")
	if err != nil {
		t.Fatalf("NewAttr failed: %v", err)
	}
	if attr == nil {
		t.Error("Attr should not be nil")
	}
	if attr.AttrId != 2 {
		t.Errorf("Expected AttrId 2, got %d", attr.AttrId)
	}
}

// TestDictionaryString tests String method
func TestDictionaryString(t *testing.T) {
	str := Builtin.String()
	if str == "" {
		t.Error("String() returned empty string")
	}
	// Should contain some attribute names
	if len(str) < 100 {
		t.Error("String() seems too short for built-in dictionary")
	}
}

// TestDictionaryMustRegister tests MustRegister method
func TestDictionaryMustRegister(t *testing.T) {
	dict := &TDictionary{}
	dict.NameItems = make(map[string]*TDictEntry)

	// Register a new attribute
	dict.MustRegister("Test-Attr", 200, AttributeText)

	// Verify it was registered
	name, ok := dict.GetName(200)
	if !ok {
		t.Error("GetName should find Test-Attr")
	}
	if name != "Test-Attr" {
		t.Errorf("Expected 'Test-Attr', got '%s'", name)
	}

	id, ok := dict.GetIndex("Test-Attr")
	if !ok {
		t.Error("GetIndex should find Test-Attr")
	}
	if id != 200 {
		t.Errorf("Expected ID 200, got %d", id)
	}
}

// TestDictionaryMustRegisterDuplicate tests MustRegister with duplicate attribute
func TestDictionaryMustRegisterDuplicate(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for duplicate attribute registration")
		}
	}()

	dict := &TDictionary{}
	dict.NameItems = make(map[string]*TDictEntry)

	dict.MustRegister("Test-Attr", 200, AttributeText)
	dict.MustRegister("Test-Attr-Dup", 200, AttributeText) // Same ID should panic
}

// TestBuiltinDictionaryAttributes tests that all built-in attributes are registered
func TestBuiltinDictionaryAttributes(t *testing.T) {
	attrs := map[byte]string{
		1:  "User-Name",
		2:  "User-Password",
		3:  "CHAP-Password",
		4:  "NAS-IP-Address",
		5:  "NAS-Port",
		6:  "Service-Type",
		7:  "Framed-Protocol",
		8:  "Framed-IP-Address",
		9:  "Framed-IP-Netmask",
		10: "Framed-Routing",
		11: "Filter-Id",
		12: "Framed-MTU",
		13: "Framed-Compression",
		14: "Login-IP-Host",
		15: "Login-Service",
		16: "Login-TCP-Port",
		18: "Reply-Message",
		19: "Callback-Number",
		20: "Callback-Id",
		22: "Framed-Route",
		23: "Framed-IPX-Network",
		24: "State",
		25: "Class",
		26: "Vendor-Specific",
		27: "Session-Timeout",
		28: "Idle-Timeout",
		29: "Termination-Action",
		30: "Called-Station-Id",
		31: "Calling-Station-Id",
		32: "NAS-Identifier",
		33: "Proxy-State",
		34: "Login-LAT-Service",
		35: "Login-LAT-Node",
		36: "Login-LAT-Group",
		37: "Framed-AppleTalk-Link",
		38: "Framed-AppleTalk-Network",
		39: "Framed-AppleTalk-Zone",
		40: "Acct-Status-Type",
		41: "Acct-Delay-Time",
		42: "Acct-Input-Octets",
		43: "Acct-Output-Octets",
		44: "Acct-Session-Id",
		45: "Acct-Authentic",
		46: "Acct-Session-Time",
		47: "Acct-Input-Packets",
		48: "Acct-Output-Packets",
		49: "Acct-Terminate-Cause",
		50: "Acct-Multi-Session-Id",
		51: "Acct-Link-Count",
		60: "CHAP-Challenge",
		61: "NAS-Port-Type",
		62: "Port-Limit",
		63: "Login-LAT-Port",
		64: "Tunnel-Type",
		65: "Tunnel-Medium-Type",
		66: "Tunnel-Client-Endpoint",
		67: "Tunnel-Server-Endpoint",
		69: "Tunnel-Password",
		76: "Authenticator-Type",
		77: "Connect-Info",
		79: "EAP-Message",
		80: "Message-Authenticator",
		87: "NAS-Port-Id",
	}

	for id, expectedName := range attrs {
		name, ok := Builtin.GetName(id)
		if !ok {
			t.Errorf("Missing attribute with ID %d", id)
			continue
		}
		if name != expectedName {
			t.Errorf("Expected name '%s' for ID %d, got '%s'", expectedName, id, name)
		}
	}
}

// TestDictionaryGetFuncForAllTypes tests GetFunc for various attribute types
func TestDictionaryGetFuncForAllTypes(t *testing.T) {
	// Text attributes
	codec := Builtin.GetFunc(1) // User-Name
	if codec.GetCodeName() != "AttributeText" {
		t.Errorf("Expected AttributeText for User-Name, got %s", codec.GetCodeName())
	}

	// String attributes
	codec = Builtin.GetFunc(3) // CHAP-Password
	if codec.GetCodeName() != "AttributeString" {
		t.Errorf("Expected AttributeString for CHAP-Password, got %s", codec.GetCodeName())
	}

	// Address attributes
	codec = Builtin.GetFunc(4) // NAS-IP-Address
	if codec.GetCodeName() != "AttributeAddress" {
		t.Errorf("Expected AttributeAddress for NAS-IP-Address, got %s", codec.GetCodeName())
	}

	// Integer attributes
	codec = Builtin.GetFunc(5) // NAS-Port
	if codec.GetCodeName() != "AttributeInteger" {
		t.Errorf("Expected AttributeInteger for NAS-Port, got %s", codec.GetCodeName())
	}

	// Vendor-Specific
	codec = Builtin.GetFunc(26) // Vendor-Specific
	if codec.GetCodeName() != "AttributeVendor" {
		t.Errorf("Expected AttributeVendor for Vendor-Specific, got %s", codec.GetCodeName())
	}

	// User-Password (RFC2865)
	codec = Builtin.GetFunc(2) // User-Password
	if codec.GetCodeName() != "RFC2865UserPassword" {
		t.Errorf("Expected RFC2865UserPassword for User-Password, got %s", codec.GetCodeName())
	}
}

// TestDictionaryNewAttrNoTransformer tests NewAttr with non-transformer attribute
func TestDictionaryNewAttrNoTransformer(t *testing.T) {
	// NAS-Port has no transformer
	attr, err := Builtin.NewAttr("NAS-Port", uint32(1234))
	if err != nil {
		t.Fatalf("NewAttr failed: %v", err)
	}
	if attr.AttrId != 5 {
		t.Errorf("Expected AttrId 5, got %d", attr.AttrId)
	}
	if attr.AttrValue.(uint32) != 1234 {
		t.Errorf("Expected 1234, got %v", attr.AttrValue)
	}
}

// TestDictionaryGetIndexMultiple tests GetIndex for multiple attributes
func TestDictionaryGetIndexMultiple(t *testing.T) {
	testCases := map[string]byte{
		"User-Name":      1,
		"User-Password":  2,
		"NAS-IP-Address": 4,
		"NAS-Port":       5,
		"Service-Type":   6,
		"Session-Timeout": 27,
	}

	for name, expectedID := range testCases {
		id, ok := Builtin.GetIndex(name)
		if !ok {
			t.Errorf("GetIndex failed for '%s'", name)
			continue
		}
		if id != expectedID {
			t.Errorf("Expected ID %d for '%s', got %d", expectedID, name, id)
		}
	}
}
