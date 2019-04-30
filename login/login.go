package login

import (
	"fmt"
	"github.com/ikuiki/go-component/language"
	"github.com/liangdas/mqant/gate"
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
	if username, ok := msg["username"]; !ok || username == "" {
		result = common.Response{
			Ret: common.RetCodeBadRequest,
			Msg: "username cannot be empty",
		}
		return
	}
	password, ok := msg["password"]
	if !ok || password == "" {
		result = common.Response{
			Ret: common.RetCodeBadRequest,
			Msg: "password cannot be empty",
		}
		return
	}
	if !m.validPassword(password.(string)) {
		result = common.Response{
			Ret: common.RetCodeBadRequest,
			Msg: "username or password incorrect",
		}
		return
	}
	username := msg["username"].(string)
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
	p, ok := m.GetModuleSettings().Settings["Password"]
	if ok {
		if pStr, ok := p.(string); ok {
			now := time.Now()
			// 密码通过日期加盐运算后作为密文传输
			timeLayout := time.RFC3339
			pwds := []string{pStr + now.Format(timeLayout)}
			// 为防止日期不一致，放宽日期到前后5分钟
			for i := 1; i < 6; i++ {
				pwds = append(pwds, pStr+now.Add(time.Duration(-i)*time.Minute).Format(timeLayout))
				pwds = append(pwds, pStr+now.Add(time.Duration(i)*time.Minute).Format(timeLayout))
			}
			// 只要密码在这11个密文中命中一个，即认为命中
			if language.ArrayIn(pwds, password) != -1 {
				return true
			}
		}
	}
	return false
}
