package login

import (
	"fmt"
	"github.com/liangdas/mqant/gate"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
	"time"
	"wegate/common"
)

func (m *Login) login(session gate.Session, msg map[string]interface{}) (result common.Response, err string) {
	if !session.IsGuest() {
		result = common.Response{
			Ret: common.RetCodeBadRequest,
			Msg: "already login",
		}
		return
	}
	username := common.ForceString(msg["username"])
	if username == "" {
		result = common.Response{
			Ret: common.RetCodeBadRequest,
			Msg: "username cannot be empty",
		}
		return
	}
	password := common.ForceString(msg["password"])
	if password == "" {
		result = common.Response{
			Ret: common.RetCodeBadRequest,
			Msg: "password cannot be empty",
		}
		return
	}
	if !m.validPassword(password) {
		result = common.Response{
			Ret: common.RetCodeBadRequest,
			Msg: "username or password incorrect",
		}
		return
	}
	err = session.Bind(username)
	if err != "" {
		return
	}
	session.Set("login", "true")
	session.Push() //推送到网关
	return common.Response{Ret: common.RetCodeOK, Msg: fmt.Sprintf("login success %s", username)}, ""
}

// totpLogin 基于时间的一次性密码登陆
func (m *Login) totpLogin(session gate.Session, msg map[string]interface{}) (result common.Response, err string) {
	if !session.IsGuest() {
		result = common.Response{
			Ret: common.RetCodeBadRequest,
			Msg: "already login",
		}
		return
	}
	username := common.ForceString(msg["username"])
	if username == "" {
		result = common.Response{
			Ret: common.RetCodeBadRequest,
			Msg: "username cannot be empty",
		}
		return
	}
	password := common.ForceString(msg["password"])
	if password == "" {
		result = common.Response{
			Ret: common.RetCodeBadRequest,
			Msg: "password cannot be empty",
		}
		return
	}
	totpCode := common.ForceString(msg["totp_code"])
	if totpCode == "" {
		result = common.Response{
			Ret: common.RetCodeBadRequest,
			Msg: "totp_code cannot be empty",
		}
		return
	}
	if !m.validTOTPPassword(password, totpCode) {
		result = common.Response{
			Ret: common.RetCodeBadRequest,
			Msg: "username or password incorrect",
		}
		return
	}
	err = session.Bind(username)
	if err != "" {
		return
	}
	session.Set("login", "true")
	session.Push() //推送到网关
	return common.Response{Ret: common.RetCodeOK, Msg: fmt.Sprintf("login success %s", username)}, ""
}

// func (m *Login) logout(session gate.Session, msg map[string]interface{}) (result common.Response, err string) {
// 	if session.IsGuest() {
// 		result = common.Response{
// 			Ret: common.RetCodeUnauthorized,
// 			Msg: "is guest, need login",
// 		}
// 		return
// 	}
// 	session.Remove("login")
// 	err = session.UnBind()
// 	if err != "" {
// 		return
// 	}
// 	err = session.Push()
// 	if err != "" {
// 		return
// 	}
// 	return common.Response{Ret: common.RetCodeOK, Msg: "logout success"}, ""
// }

func (m *Login) validPassword(password string) (valid bool) {
	if m.password != "" {
		now := time.Now()
		// 密码通过日期加盐运算后作为密文传输
		timeLayout := time.RFC822
		pwds := []string{m.password + now.Format(timeLayout)}
		// 为防止日期不一致，放宽日期到前后5分钟
		for i := 1; i < 6; i++ {
			pwds = append(pwds, m.password+now.Add(time.Duration(-i)*time.Minute).Format(timeLayout))
			pwds = append(pwds, m.password+now.Add(time.Duration(i)*time.Minute).Format(timeLayout))
		}
		// 只要密码在这11个密文中命中一个，即认为命中
		for _, pwd := range pwds {
			if bcrypt.CompareHashAndPassword([]byte(password), []byte(pwd)) == nil {
				return true
			}
		}
	}
	return false
}

func (m *Login) validTOTPPassword(password, totpCode string) (valid bool) {
	if m.totpPasswd != "" && m.totpSecret != "" {
		now := time.Now()
		// 密码通过日期加盐运算后作为密文传输
		timeLayout := time.RFC822
		pwds := []string{m.totpPasswd + now.Format(timeLayout)}
		// 为防止日期不一致，放宽日期到前后5分钟
		for i := 1; i < 6; i++ {
			pwds = append(pwds, m.totpPasswd+now.Add(time.Duration(-i)*time.Minute).Format(timeLayout))
			pwds = append(pwds, m.totpPasswd+now.Add(time.Duration(i)*time.Minute).Format(timeLayout))
		}
		// 只要密码在这11个密文中命中一个，即认为命中
		for _, pwd := range pwds {
			if bcrypt.CompareHashAndPassword([]byte(password), []byte(pwd)) == nil {
				// 密码命中，核对otpCode
				pass, _ := totp.ValidateCustom(totpCode, m.totpSecret, time.Now(), totp.ValidateOpts{
					Period: m.totpPeriod,
					Digits: m.totpDigits,
				})
				return pass
			}
		}
	}
	return false
}
