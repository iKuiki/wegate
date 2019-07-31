package login

/**
一定要记得在confin.json配置这个模块的参数,否则无法使用
*/

import (
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/module/base"
	"github.com/pquerna/otp"
)

// Module 模块实例化
func Module() module.Module {
	m := new(Login)
	listener := new(Listener)
	m.SetListener(listener)
	return m
}

// Login 登陆模块
type Login struct {
	basemodule.BaseModule
	password   string
	totpPasswd string
	totpSecret string
	totpPeriod uint
	totpDigits otp.Digits
}

// GetType 获取模块类型
func (m *Login) GetType() string {
	//很关键,需要与配置文件中的Module配置对应
	return "Login"
}

// Version 获取模块Version
func (m *Login) Version() string {
	//可以在监控时了解代码版本
	return "1.0.0"
}

// OnInit 模块初始化
func (m *Login) OnInit(app module.App, settings *conf.ModuleSettings) {
	m.BaseModule.OnInit(m, app, settings)
	// 我们约定所有对客户端的请求都以Handler_开头
	m.GetServer().RegisterGO("HD_Login", m.login)         // 登陆
	m.GetServer().RegisterGO("HD_TOTPLogin", m.totpLogin) // 基于时间的一次性密码登陆
	// m.GetServer().RegisterGO("HD_Logout", m.logout) // 注销

	// 处理配置
	if p, ok := m.GetModuleSettings().Settings["Password"]; ok {
		if pStr, ok := p.(string); ok {
			m.password = pStr
		}
	}
	if p, ok := m.GetModuleSettings().Settings["TOTPPasswd"]; ok {
		if pStr, ok := p.(string); ok {
			m.totpPasswd = pStr
		}
	}
	if p, ok := m.GetModuleSettings().Settings["TOTPSecret"]; ok {
		if pStr, ok := p.(string); ok {
			m.totpSecret = pStr
		}
	}
	if m.totpPasswd != "" && m.totpSecret != "" {
		log.Info("TOTP settings found, TOTP login enabled")
		if p, ok := m.GetModuleSettings().Settings["TOTPPeriod"]; ok {
			if pNum, ok := p.(float64); ok {
				m.totpPeriod = uint(pNum)
			}
		}
		if m.totpPeriod <= 0 {
			log.Info("TOTP Period setting invalid, use default 30")
			m.totpPeriod = 30
		}
		if p, ok := m.GetModuleSettings().Settings["TOTPDigits"]; ok {
			if pNum, ok := p.(float64); ok {
				switch pNum {
				case 6:
					m.totpDigits = otp.DigitsSix
				case 8:
					m.totpDigits = otp.DigitsEight
				default:
					log.Warning("unknown TOTP Digits: %d, use default: 8", pNum)
					m.totpDigits = otp.DigitsEight
				}
			}
		}
		if m.totpDigits == 0 {
			log.Info("TOTP Digits setting missing, use default 8")
			m.totpDigits = otp.DigitsEight
		}
	}
}

// Run 运行主函数
func (m *Login) Run(closeSig chan bool) {
	for {
		select {
		case <-closeSig:
			return
		}
	}
}

// OnDestroy 析构函数
func (m *Login) OnDestroy() {
	//一定别忘了关闭RPC
	m.GetServer().OnDestroy()
}
