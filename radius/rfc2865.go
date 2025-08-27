package radius

import (
	"bytes"
	"crypto/md5"
	"errors"
)

func init() {
	builtinOnce.Do(initDictionary)
	// TODO: Attribute* should be initialized before
	Builtin.MustRegister("User-Name", 1, AttributeText)
	Builtin.MustRegister("User-Password", 2, rfc2865UserPassword{})
	Builtin.MustRegister("CHAP-Password", 3, AttributeString)
	Builtin.MustRegister("NAS-IP-Address", 4, AttributeAddress)
	Builtin.MustRegister("NAS-Port", 5, AttributeInteger)
	Builtin.MustRegister("Service-Type", 6, AttributeInteger)
	Builtin.MustRegister("Framed-Protocol", 7, AttributeInteger)
	Builtin.MustRegister("Framed-IP-Address", 8, AttributeAddress)
	Builtin.MustRegister("Framed-IP-Netmask", 9, AttributeAddress)
	Builtin.MustRegister("Framed-Routing", 10, AttributeInteger)
	Builtin.MustRegister("Filter-Id", 11, AttributeText)
	Builtin.MustRegister("Framed-MTU", 12, AttributeInteger)
	Builtin.MustRegister("Framed-Compression", 13, AttributeInteger)
	Builtin.MustRegister("Login-IP-Host", 14, AttributeAddress)
	Builtin.MustRegister("Login-Service", 15, AttributeInteger)
	Builtin.MustRegister("Login-TCP-Port", 16, AttributeInteger)
	Builtin.MustRegister("Reply-Message", 18, AttributeText)
	Builtin.MustRegister("Callback-Number", 19, AttributeString)
	Builtin.MustRegister("Callback-Id", 20, AttributeString)
	Builtin.MustRegister("Framed-Route", 22, AttributeText)
	Builtin.MustRegister("Framed-IPX-Network", 23, AttributeAddress)
	Builtin.MustRegister("State", 24, AttributeString)
	Builtin.MustRegister("Class", 25, AttributeString)
	Builtin.MustRegister("Vendor-Specific", 26, AttributeVendor)
	//Builtin.MustRegister("Vendor-Specific", 26, AttributeText)
	Builtin.MustRegister("Session-Timeout", 27, AttributeInteger)
	Builtin.MustRegister("Idle-Timeout", 28, AttributeInteger)
	Builtin.MustRegister("Termination-Action", 29, AttributeInteger)
	Builtin.MustRegister("Called-Station-Id", 30, AttributeString)
	Builtin.MustRegister("Calling-Station-Id", 31, AttributeString)
	Builtin.MustRegister("NAS-Identifier", 32, AttributeString)
	Builtin.MustRegister("Proxy-State", 33, AttributeString)
	Builtin.MustRegister("Login-LAT-Service", 34, AttributeString)
	Builtin.MustRegister("Login-LAT-Node", 35, AttributeString)
	Builtin.MustRegister("Login-LAT-Group", 36, AttributeString)
	Builtin.MustRegister("Framed-AppleTalk-Link", 37, AttributeInteger)
	Builtin.MustRegister("Framed-AppleTalk-Network", 38, AttributeInteger)
	Builtin.MustRegister("Framed-AppleTalk-Zone", 39, AttributeString)
	Builtin.MustRegister("CHAP-Challenge", 60, AttributeString)
	Builtin.MustRegister("NAS-Port-Type", 61, AttributeInteger)
	Builtin.MustRegister("Port-Limit", 62, AttributeInteger)
	Builtin.MustRegister("Login-LAT-Port", 63, AttributeString)

	Builtin.MustRegister("Tunnel-Type", 64, AttributeInteger)
	Builtin.MustRegister("Tunnel-Medium-Type", 65, AttributeInteger)
	Builtin.MustRegister("Tunnel-Client-Endpoint", 66, AttributeString)
	Builtin.MustRegister("Tunnel-Server-Endpoint", 67, AttributeString)
	Builtin.MustRegister("Tunnel-Password", 69, AttributeString)

	// FreeRADIUS specific
	Builtin.MustRegister("Authenticator-Type", 76, AttributeString)
	Builtin.MustRegister("Connect-Info", 77, AttributeString)
	Builtin.MustRegister("EAP-Message", 79, AttributeString)
	Builtin.MustRegister("Message-Authenticator", 80, AttributeString)
	Builtin.MustRegister("NAS-Port-Id", 87, AttributeString)
	Builtin.MustRegister("FreeRADIUS-Statistics-Type", 127, AttributeInteger)

	Builtin.MustRegister("FreeRADIUS-Total-Access-Requests", 128, AttributeString)
	Builtin.MustRegister("FreeRADIUS-Total-Access-Accepts", 129, AttributeString)
	Builtin.MustRegister("FreeRADIUS-Total-Access-Rejects", 130, AttributeString)
	Builtin.MustRegister("FreeRADIUS-Total-Access-Challenges", 131, AttributeString)
	Builtin.MustRegister("FreeRADIUS-Total-Auth-Responses", 132, AttributeString)
	Builtin.MustRegister("FreeRADIUS-Total-Auth-Duplicate-Requests", 133, AttributeString)
	Builtin.MustRegister("FreeRADIUS-Total-Auth-Malformed-Requests", 134, AttributeString)
	Builtin.MustRegister("FreeRADIUS-Total-Auth-Invalid-Requests", 135, AttributeString)
	Builtin.MustRegister("FreeRADIUS-Total-Auth-Dropped-Requests", 136, AttributeString)
	Builtin.MustRegister("FreeRADIUS-Total-Auth-Unknown-Types", 137, AttributeString)
}

type rfc2865UserPassword struct{}

func (rfc2865UserPassword) Decode(p *TDataPacket, value []byte) (interface{}, error) {
	if p.Secret == nil {
		return nil, errors.New("radius: User-Password attribute requires Packet.Secret")
	}
	if len(value) != 16 {
		return nil, errors.New("radius: invalid User-Password attribute length")
	}
	v := make([]byte, len(value))
	copy(v, value)

	var mask [md5.Size]byte
	hash := md5.New()
	hash.Write(p.Secret)
	hash.Write(p.Authenticator[:])
	hash.Sum(mask[0:0])

	for i, b := range v {
		v[i] = b ^ mask[i]
	}

	if i := bytes.IndexByte(v, 0); i > -1 {
		return string(v[:i]), nil
	}
	return string(v), nil
}

func (rfc2865UserPassword) Encode(p *TDataPacket, value interface{}) ([]byte, error) {
	if p.Secret == nil {
		return nil, errors.New("radius: User-Password attribute requires Packet.Secret")
	}
	var password []byte
	if bytePassword, ok := value.([]byte); !ok {
		strPassword, ok := value.(string)
		if !ok {
			return nil, errors.New("radius: User-Password attribute must be string or []byte")
		}
		password = []byte(strPassword)
	} else {
		password = bytePassword
	}

	if len(password) > 16 {
		return nil, errors.New("radius: invalid User-Password attribute length")
	}

	var mask [md5.Size]byte
	hash := md5.New()
	hash.Write(p.Secret)
	hash.Write(p.Authenticator[:])
	hash.Sum(mask[0:0])

	for i, b := range password {
		mask[i] = b ^ mask[i]
	}

	return mask[:], nil
}
func (rfc2865UserPassword) GetCodeName() string {
	return "RFC2865UserPassword"
}
