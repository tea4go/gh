package ldapclient

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	ldap "github.com/go-ldap/ldap/v3"
	"github.com/mozillazg/go-pinyin"
	"github.com/tea4go/gh/utils"
	"golang.org/x/text/encoding/unicode"
)

/*
https://github.com/opaas/winad-client-go
https://github.com/antoineduban/ldap
https://github.com/thammuio/ldap-passwd-webui
https://github.com/github-9527/ohedu // ChangePassword
*/

var DefaultPwd string = "BGgsh*1050"

var uacMask = map[uint32]string{
	2:        "ACCOUNT_DISABLE",                        //禁用用户帐户
	8:        "HOMEDIR_REQUIRED",                       //主文件夹是必需的
	16:       "LOCKOUT",                                //
	32:       "PASSWD_NOTREQD",                         //不需要密码
	64:       "PASSWD_CANT_CHANGE",                     //用户不能更改密码。可以读取此标志，但不能直接设置它
	128:      "ENCRYPTED_TEXT_PASSWORD_ALLOWED",        //用户可以发送加密的密码
	512:      "NORMAL_ACCOUNT",                         //表示典型用户的默认帐户类型
	8192:     "SERVER_TRUST_ACCOUNT",                   //
	65536:    "DONT_EXPIRE_PASSWD",                     //该帐户上永远不会过期的密码。
	131072:   "MNS_LOGON_ACCOUNT",                      //
	262144:   "SMARTCARD_REQUIRED",                     //
	524288:   "TRUSTED_FOR_DELEGATION",                 //
	1048576:  "NOT_DELEGATED",                          //
	2097152:  "USE_DES_KEY_ONLY",                       //
	4194304:  "DONT_REQUIRE_PREAUTH",                   //
	8388608:  "PASSWORD_EXPIRED",                       //用户的密码已过期
	16777216: "TRUSTED_TO_AUTHENTICATE_FOR_DELEGATION", //
	33554432: "NO_AUTH_DATA_REQUIRED",                  //
	67108864: "PARTIAL_SECRETS_ACCOUNT",                //
}

type TResultLdap struct {
	DN    string              `json:"dn"`
	Attrs map[string][]string `json:"attributes"`
}

func (Self *TResultLdap) GetAttr(name string) string {
	if v, ok := Self.Attrs[name]; ok {
		if len(v) == 1 {
			return v[0]
		} else {
			return strings.Join(v, ";")
		}
	} else {
		return ""
	}
}

func (Self *TResultLdap) GetAttrByInt(name string) int {
	if v, ok := Self.Attrs[name]; ok {
		if len(v) == 1 {
			vi, _ := strconv.Atoi(v[0])
			return vi
		} else {
			return -1
		}
	} else {
		return -1
	}
}

func (Self *TResultLdap) String() string {
	s, _ := json.Marshal(Self)
	var out bytes.Buffer
	json.Indent(&out, s, "", "\t")
	return string(out.Bytes())
}

type TLdapUser struct {
	StaffCode string `json:"code"`    //编码
	StaffName string `json:"name"`    //姓名
	Email     string `json:"mail"`    //邮箱
	Org       string `json:"org"`     //组织
	Dept      string `json:"dept"`    //部门
	Phone     string `json:"phone"`   //手机
	Company   string `json:"belong"`  //公司
	Station   string `json:"station"` //属地
}

func (Self *TLdapUser) String() string {
	s, _ := json.Marshal(Self)
	var out bytes.Buffer
	json.Indent(&out, s, "", "\t")
	return string(out.Bytes())
}

type TLdapClient struct {
	Addr       string   `json:"addr"`
	BaseDn     string   `json:"baseDn"`
	BindDn     string   `json:"bindDn`
	BindPass   string   `json:"bindPass"`
	AuthFilter string   `json:"authFilter"`
	Attributes []string `json:"attributes"`
	MailDomain string   `json:"mailDomain"`
	TLS        bool     `json:"tls"`
	StartTLS   bool     `json:"startTLS"`
	Conn       *ldap.Conn
}

func (lc *TLdapClient) Close() {
	if lc.Conn != nil {
		lc.Conn.Close()
		lc.Conn = nil
	}
}

func (lc *TLdapClient) Connect() (err error) {
	if lc.TLS {
		lc.Conn, err = ldap.DialTLS("tcp", lc.Addr, &tls.Config{InsecureSkipVerify: true})
	} else {
		lc.Conn, err = ldap.Dial("tcp", lc.Addr)
	}
	if err != nil {
		return err
	}
	//lc.Conn.Debug = true

	if !lc.TLS && lc.StartTLS {
		err = lc.Conn.StartTLS(&tls.Config{InsecureSkipVerify: true})
		if err != nil {
			lc.Conn.Close()
			lc.Conn = nil
			return err
		}
	}

	err = lc.Conn.Bind(lc.BindDn, lc.BindPass)
	if err != nil {
		lc.Conn.Close()
		lc.Conn = nil
		return err
	}
	return nil
}

func (lc *TLdapClient) IsClosing() bool {
	if lc.Conn == nil || lc.Conn.IsClosing() {
		return true
	}
	return false
}

func (lc *TLdapClient) Bind(username, password string) (success bool, err error) {
	if lc.IsClosing() {
		return false, fmt.Errorf("服务器未连接，请先连接服务器！")
	}
	err = lc.Conn.Bind(username, password)
	if err != nil {
		return
	}
	success = true
	return
}

func (lc *TLdapClient) Auth(username, password string) (success bool, err error) {
	if lc.IsClosing() {
		return false, fmt.Errorf("服务器未连接，请先连接服务器！")
	}
	searchRequest := ldap.NewSearchRequest(
		lc.BaseDn,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf(lc.AuthFilter, username),
		lc.Attributes,
		nil,
	)
	sr, err := lc.Conn.Search(searchRequest)
	if err != nil {
		return
	}
	if len(sr.Entries) == 0 {
		err = errors.New("没有找到用户！")
		return
	}
	if len(sr.Entries) > 1 {
		err = errors.New("用户不唯一！")
		return
	}

	err = lc.Conn.Bind(sr.Entries[0].DN, password)
	if err != nil {
		return
	}

	//重新绑定为搜索用户以进行任何其他查询
	err = lc.Conn.Bind(lc.BindDn, lc.BindPass)
	if err != nil {
		return
	}
	success = true
	return
}

func (lc *TLdapClient) GetSearchResult(entry *ldap.Entry) *TResultLdap {
	var user TResultLdap
	attributes := make(map[string][]string)
	for _, attr := range entry.Attributes {
		if attr.Name == "objectGUID" {
			var b [16]byte
			copy(b[:], attr.ByteValues[0])
			value := utils.GUIDFromWindowsArray(b)
			attributes[attr.Name] = []string{value.String()}
		} else if attr.Name == "pwdLastSet" || attr.Name == "badPasswordTime" || attr.Name == "lastLogon" || attr.Name == "lockoutTime" {
			// lastLogon  用户上次登录的时间(每次用户登录时)
			// badPasswordTime 上次使用错误密码的时间
			// pwdLastSet 指定用户上次设置密码时间，若要强制用户在下次登录时更改其密码，请将 pwdLastSet 属性设置为零 (0) 。 若要删除此要求，请将 pwdLastSet 属性设置为 -1。 pwdLastSet 属性不能设置为除系统以外的任何其他值。
			t, _ := ParseTicks(attr.Values[0])
			attributes[attr.Name] = []string{t.Local().Format(utils.DateTimeFormat)}
		} else if attr.Name == "whenCreated" || attr.Name == "whenChanged" {
			// whenCreated 创建此对象的日期
			// whenChanged 上次更改此对象的日期
			t, _ := time.Parse("20060102150405.0Z0700", string(attr.Values[0]))
			attributes[attr.Name] = []string{t.Local().Format(utils.DateTimeFormat)}
		} else {
			attributes[attr.Name] = attr.Values
		}
	}

	user.DN = entry.DN
	user.Attrs = attributes
	return &user
}

func (lc *TLdapClient) SearchUser(username string) (user *TResultLdap, err error) {
	if lc.IsClosing() {
		return nil, fmt.Errorf("服务器未连接，请先连接服务器！")
	}
	searchRequest := ldap.NewSearchRequest(
		lc.BaseDn,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf(lc.AuthFilter, username),
		lc.Attributes,
		nil,
	)
	//fmt.Println(utils.GetJson(searchRequest))
	sr, err := lc.Conn.Search(searchRequest)
	if err != nil {
		return
	}
	if len(sr.Entries) == 0 {
		err = errors.New("没有找到用户！")
		return
	}
	if len(sr.Entries) > 1 {
		err = errors.New("用户不唯一！")
		return
	}
	user = lc.GetSearchResult(sr.Entries[0])
	return
}

func (lc *TLdapClient) Search(SearchFilter string, scope int) (results []*TResultLdap, err error) {
	if lc.IsClosing() {
		return nil, fmt.Errorf("服务器未连接，请先连接服务器！")
	}
	searchRequest := ldap.NewSearchRequest(
		lc.BaseDn,
		scope,
		ldap.NeverDerefAliases, 0, 0, false,
		SearchFilter,
		lc.Attributes,
		nil,
	)
	//fmt.Println(utils.GetJson(searchRequest))
	sr, err := lc.Conn.Search(searchRequest)
	if err != nil {
		return
	}
	if len(sr.Entries) == 0 {
		err = errors.New("没有找到记录！")
		return
	}

	results = []*TResultLdap{}
	var result *TResultLdap
	for _, entry := range sr.Entries {
		result = lc.GetSearchResult(entry)
		results = append(results, result)
	}
	return
}

func (lc *TLdapClient) SearchBase(SearchFilter string) (results []*TResultLdap, err error) {
	if lc.IsClosing() {
		return nil, fmt.Errorf("服务器未连接，请先连接服务器！")
	}
	searchRequest := ldap.NewSearchRequest(
		lc.BaseDn,
		ldap.ScopeBaseObject,
		ldap.NeverDerefAliases, 0, 0, false,
		SearchFilter,
		lc.Attributes,
		nil,
	)
	//fmt.Println(utils.GetJson(searchRequest))
	sr, err := lc.Conn.Search(searchRequest)
	if err != nil {
		return
	}
	if len(sr.Entries) == 0 {
		err = errors.New("没有找到记录！")
		return
	}

	results = []*TResultLdap{}
	var result *TResultLdap
	for _, entry := range sr.Entries {
		result = lc.GetSearchResult(entry)
		results = append(results, result)
	}
	return
}

func (lc *TLdapClient) SearchSubOne(SearchFilter string) (results []*TResultLdap, err error) {
	if lc.IsClosing() {
		return nil, fmt.Errorf("服务器未连接，请先连接服务器！")
	}
	searchRequest := ldap.NewSearchRequest(
		lc.BaseDn,
		ldap.ScopeSingleLevel,
		ldap.NeverDerefAliases, 0, 0, false,
		SearchFilter,
		lc.Attributes,
		nil,
	)
	//fmt.Println(utils.GetJson(searchRequest))
	sr, err := lc.Conn.Search(searchRequest)
	if err != nil {
		return
	}
	if len(sr.Entries) == 0 {
		err = errors.New("没有找到记录！")
		return
	}

	results = []*TResultLdap{}
	var result *TResultLdap
	for _, entry := range sr.Entries {
		result = lc.GetSearchResult(entry)
		results = append(results, result)
	}
	return
}

func (lc *TLdapClient) SearchSubAll(SearchFilter string) (results []*TResultLdap, err error) {
	if lc.IsClosing() {
		return nil, fmt.Errorf("服务器未连接，请先连接服务器！")
	}
	searchRequest := ldap.NewSearchRequest(
		lc.BaseDn,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases, 0, 0, false,
		SearchFilter,
		lc.Attributes,
		nil,
	)
	sr, err := lc.Conn.Search(searchRequest)
	if err != nil {
		return
	}
	if len(sr.Entries) == 0 {
		err = errors.New("没有找到记录！")
		return
	}

	results = []*TResultLdap{}
	var result *TResultLdap
	for _, entry := range sr.Entries {
		result = lc.GetSearchResult(entry)
		results = append(results, result)
	}
	return
}

func (lc *TLdapClient) SetExternalEmailAddress(path, staff_code, mail, display_name string) error {
	if lc.IsClosing() {
		return fmt.Errorf("服务器未连接，请先连接服务器！")
	}
	if path != "" {
		path = path + ","
	}
	dn := fmt.Sprintf("CN=%s,%s%s", staff_code, path, lc.BaseDn)

	modRequest := ldap.NewModifyRequest(dn, nil)
	modRequest.Replace("mail", []string{mail})
	modRequest.Replace("mailNickname", []string{display_name})
	modRequest.Replace("proxyAddresses", []string{
		"SMTP:" + mail,
		"smtp:" + staff_code + lc.MailDomain,
	})
	return lc.Conn.Modify(modRequest)
}

// https://learn.microsoft.com/zh-cn/windows/win32/ad/creating-a-user
func (lc *TLdapClient) CreateUser(path string, ldapUser *TLdapUser) (string, string, error) {
	if lc.IsClosing() {
		return "", "", fmt.Errorf("服务器未连接，请先连接服务器！")
	}
	if path != "" {
		path = path + ","
	}
	dn := fmt.Sprintf("CN=%s,%s%s", ldapUser.StaffCode, path, lc.BaseDn)
	addRequest := ldap.NewAddRequest(dn, nil)

	userAccountControl := "544"
	pwdLastSet := "0"
	mailDomain := lc.MailDomain
	//fmt.Println(ldapUser.Email, userAccountControl, pwdLastSet, DefaultPwd, mailDomain)
	if nn := strings.Split(ldapUser.Email, "@"); len(nn) == 2 {
		mailDomain = "@" + nn[1]
	} else {
		ldapUser.Email = ldapUser.Email + mailDomain
	}
	//fmt.Println("===>", ldapUser.Email)
	displayName := ldapUser.StaffName + strings.Join(pinyin.LazyConvert(ldapUser.StaffName, nil), "") + ldapUser.StaffCode
	mailNickname := strings.Join(pinyin.LazyConvert(ldapUser.StaffName, nil), "")
	//fmt.Println("===>", displayName)
	//fmt.Println("===>", mailNickname)

	// default attributes for adop user
	addRequest.Attribute("objectClass", []string{"user", "organizationalPerson", "person", "top"})

	// assign values
	addRequest.Attribute("cn", []string{ldapUser.StaffCode})
	addRequest.Attribute("sAMAccountName", []string{ldapUser.StaffCode})
	addRequest.Attribute("name", []string{ldapUser.StaffCode})
	addRequest.Attribute("sn", []string{ldapUser.StaffName})
	addRequest.Attribute("displayName", []string{displayName})
	addRequest.Attribute("userPrincipalName", []string{ldapUser.StaffCode + mailDomain})
	//addRequest.Attribute("mail", []string{ldapUser.Email})
	//addRequest.Attribute("mailNickname", []string{mailNickname})
	addRequest.Attribute("description", []string{ldapUser.Org})
	addRequest.Attribute("department", []string{ldapUser.Dept})
	addRequest.Attribute("telephoneNumber", []string{ldapUser.Phone})
	addRequest.Attribute("mobile", []string{ldapUser.Phone})
	addRequest.Attribute("company", []string{ldapUser.Company})
	addRequest.Attribute("l", []string{ldapUser.Station})

	addRequest.Attribute("proxyAddresses", []string{
		"SMTP:" + ldapUser.Email,
		"smtp:" + ldapUser.StaffCode + mailDomain,
	})

	/*
		accountExpires	指定帐户何时过期。 默认值 为TIMEQ_FOREVER，指示帐户永远不会过期。
		userAccountControl	包含确定用户多个登录和帐户功能的值。
		默认情况下，将设置以下标志：
		UF_ACCOUNTDISABLE - 帐户已禁用。
		UF_PASSWD_NOTREQD - 不需要密码。
		UF_NORMAL_ACCOUNT - 表示典型用户的默认帐户类型。
	*/
	addRequest.Attribute("userAccountControl", []string{userAccountControl})
	addRequest.Attribute("pwdLastSet", []string{pwdLastSet})

	if DefaultPwd != "" {
		utf16 := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
		newEncoded, err := utf16.NewEncoder().String(fmt.Sprintf(`"%s"`, DefaultPwd))
		if err == nil {
			addRequest.Attribute("unicodePwd", []string{newEncoded})
		}
	}

	// Add user
	return dn, mailNickname, lc.Conn.Add(addRequest)
}

func (lc *TLdapClient) ChangePassword(path string, staff_code, new_password string) error {
	if lc.IsClosing() {
		return fmt.Errorf("服务器未连接，请先连接服务器！")
	}
	if path != "" {
		path = path + ","
	}
	dn := fmt.Sprintf("CN=%s,%s%s", staff_code, path, lc.BaseDn)

	utf16 := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	newEncoded, err := utf16.NewEncoder().String(fmt.Sprintf(`"%s"`, new_password))
	if err != nil {
		return fmt.Errorf("密码编码失败，%s", err.Error())
	}

	modRequest := ldap.NewModifyRequest(dn, nil)
	modRequest.Replace("unicodePwd", []string{string(newEncoded)})
	return lc.Conn.Modify(modRequest)
}

func (lc *TLdapClient) EnableAccount(staff_code string) error {
	if lc.IsClosing() {
		return fmt.Errorf("服务器未连接，请先连接服务器！")
	}
	user, err := lc.SearchUser(staff_code)
	if err != nil {
		return fmt.Errorf("查找用户失败，%s", err.Error())
	}
	newAC, err := strconv.Atoi(user.Attrs["userAccountControl"][0])
	if err != nil {
		return fmt.Errorf("用户[userAccountControl]属性格式错误，不是整数！")
	}
	newAC = newAC | 0x10000 //设置账号密码不过期
	newAC = newAC & ^0x0002 //启用账号

	dn := user.GetAttr("distinguishedName")
	modRequest := ldap.NewModifyRequest(dn, nil)
	modRequest.Replace("userAccountControl", []string{fmt.Sprintf("%d", newAC)})
	modRequest.Replace("lockOutTime", []string{"0"})
	return lc.Conn.Modify(modRequest)
}

func (lc *TLdapClient) DisableAccount(staff_code string) error {
	if lc.IsClosing() {
		return fmt.Errorf("服务器未连接，请先连接服务器！")
	}
	user, err := lc.SearchUser(staff_code)
	if err != nil {
		return fmt.Errorf("查找用户失败，%s", err.Error())
	}
	newAC, err := strconv.Atoi(user.Attrs["userAccountControl"][0])
	if err != nil {
		return fmt.Errorf("用户[userAccountControl]属性格式错误，不是整数！")
	}
	newAC = newAC | 0x0002

	dn := user.GetAttr("distinguishedName")
	modRequest := ldap.NewModifyRequest(dn, nil)
	modRequest.Replace("userAccountControl", []string{fmt.Sprintf("%d", newAC)})
	return lc.Conn.Modify(modRequest)
}

func (lc *TLdapClient) DeleteUser(path string, staff_code string) error {
	if lc.IsClosing() {
		return fmt.Errorf("服务器未连接，请先连接服务器！")
	}
	if path != "" {
		path = path + ","
	}
	dn := fmt.Sprintf("CN=%s,%s%s", staff_code, path, lc.BaseDn)
	deleteRequest := ldap.NewDelRequest(dn, nil)

	// Delete User
	return lc.Conn.Del(deleteRequest)
}

func (lc *TLdapClient) CreatePath(path string, pathName string) error {
	if lc.IsClosing() {
		return fmt.Errorf("服务器未连接，请先连接服务器！")
	}
	if path != "" {
		path = path + ","
	}
	dn := fmt.Sprintf("OU=%s,%s%s", pathName, path, lc.BaseDn)
	addRequest := ldap.NewAddRequest(dn, nil)

	// default attributes for adop path
	addRequest.Attribute("objectClass", []string{"top", "organizationalUnit"})
	addRequest.Attribute("instanceType", []string{fmt.Sprintf("%d", 0x00000004)})

	// assign values
	addRequest.Attribute("ou", []string{pathName})
	addRequest.Attribute("name", []string{pathName})

	// Add path
	return lc.Conn.Add(addRequest)
}

func (lc *TLdapClient) DeletePath(path string, path_name string) error {
	if lc.IsClosing() {
		return fmt.Errorf("服务器未连接，请先连接服务器！")
	}
	if path != "" {
		path = path + ","
	}
	dn := fmt.Sprintf("OU=%s,%s%s", path_name, path, lc.BaseDn)
	deleteRequest := ldap.NewDelRequest(dn, nil)

	// Delete User
	return lc.Conn.Del(deleteRequest)
}

func (lc *TLdapClient) CreateGroup(path string, groupName string) error {
	if lc.IsClosing() {
		return fmt.Errorf("服务器未连接，请先连接服务器！")
	}
	if path != "" {
		path = path + ","
	}
	dn := fmt.Sprintf("CN=%s,%s%s", groupName, path, lc.BaseDn)
	addRequest := ldap.NewAddRequest(dn, nil)

	// default attributes for adop group
	addRequest.Attribute("objectClass", []string{"top", "group"})
	addRequest.Attribute("instanceType", []string{fmt.Sprintf("%d", 0x00000004)})
	addRequest.Attribute("groupType", []string{fmt.Sprintf("%d", 0x80000002)})

	// assign values
	addRequest.Attribute("cn", []string{groupName})
	addRequest.Attribute("sAMAccountName", []string{groupName})
	addRequest.Attribute("name", []string{groupName})

	// Add group
	return lc.Conn.Add(addRequest)
}

func (lc *TLdapClient) DeleteGroup(path string, group_name string) error {
	if lc.IsClosing() {
		return fmt.Errorf("服务器未连接，请先连接服务器！")
	}
	if path != "" {
		path = path + ","
	}
	dn := fmt.Sprintf("CN=%s,%s%s", group_name, path, lc.BaseDn)
	deleteRequest := ldap.NewDelRequest(dn, nil)

	// Delete Group
	return lc.Conn.Del(deleteRequest)
}

func (lc *TLdapClient) AddGroupUser(path string, groupName, userDN string) error {
	if lc.IsClosing() {
		return fmt.Errorf("服务器未连接，请先连接服务器！")
	}
	if path != "" {
		path = path + ","
	}
	dn := fmt.Sprintf("CN=%s,%s%s", groupName, path, lc.BaseDn)
	modRequest := ldap.NewModifyRequest(dn, nil)
	modRequest.Add("member", []string{userDN + "," + lc.BaseDn})
	return lc.Conn.Modify(modRequest)
}

func (lc *TLdapClient) DelGroupUser(path string, groupName, userDN string) error {
	if lc.IsClosing() {
		return fmt.Errorf("服务器未连接，请先连接服务器！")
	}
	if path != "" {
		path = path + ","
	}
	dn := fmt.Sprintf("CN=%s,%s%s", groupName, path, lc.BaseDn)
	modRequest := ldap.NewModifyRequest(dn, nil)
	modRequest.Delete("member", []string{userDN + "," + lc.BaseDn})
	return lc.Conn.Modify(modRequest)
}

func init() {
	ldap.Logger(log.Default())
}
