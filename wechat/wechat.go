package wechat

import (
	"fmt"
	"github.com/ikuiki/wwdk"
	"github.com/ikuiki/wwdk/datastruct"
	"github.com/liangdas/mqant/log"
	"time"
	"wegate/wechat/wechatstruct"
)

func (m *Wechat) wechatServe(closeSig <-chan bool) {
	// 创建登陆用channel用于回传登陆信息
	loginChan := make(chan wwdk.LoginChannelItem)
	m.wechat.Login(loginChan)
	// 新建一个func根据channel返回信息进行处理
LOGINLOOP:
	for {
		select {
		case item, ok := <-loginChan:
			if ok {
				m.updateLoginStatus(item)
			} else {
				break LOGINLOOP
			}
		case <-closeSig:
			return
		}
	}
	// 创建同步channel
	syncChannel := make(chan wwdk.SyncChannelItem)
	// 将channel传入startServe方法，开始同步服务并且将新信息通过syncChannel传回
	m.wechat.StartServe(syncChannel)
	// 新建一个func处理syncChannel传回信息
	// 新建一个方法，是为了能够方便的return
	// 之所以用select不用for range，是为了处理closeSig
SYNCLOOP:
	for {
		select {
		case item, ok := <-syncChannel:
			if ok {
				// 在子方法内执行逻辑
				switch item.Code {
				// 联系人变更
				case wwdk.SyncStatusModifyContact:
					contact := *item.Contact
					headImg, err := m.wechat.SaveContactImg(contact)
					if err == nil {
						contact.HeadImgURL = headImg
					}
					m.contacts[contact.UserName] = contact
					// 广播contact
					m.broadcastContact(contact)
				// 收到新信息
				case wwdk.SyncStatusNewMessage:
					// 如果为媒体消息，则下载媒体
					message := *item.Message
					switch message.MsgType {
					case datastruct.ImageMsg:
						if fileurl, err := m.wechat.SaveMessageImage(message); err != nil {
							log.Debug("wechat.SaveMessageImage error: %v", err)
						} else {
							message.FileName = fileurl
						}
					case datastruct.VoiceMsg:
						if fileurl, err := m.wechat.SaveMessageVoice(message); err != nil {
							log.Debug("wechat.SaveMessageImage error: %v", err)
						} else {
							message.FileName = fileurl
						}
					case datastruct.LittleVideoMsg:
						if fileurl, err := m.wechat.SaveMessageVideo(message); err != nil {
							log.Debug("wechat.SaveMessageImage error: %v", err)
						} else {
							message.FileName = fileurl
						}
					}
					// 广播message
					m.broadcastMessage(message)
				case wwdk.SyncStatusPanic:
					// 发生致命错误，sync中断
					panic(fmt.Sprintf("sync panic: %+v", item.Err))
				}
			} else {
				break SYNCLOOP
			}
		case <-closeSig:
			return
		}
	}
}

// broadcastContact 广播contact
func (m *Wechat) broadcastContact(contact datastruct.Contact) {
	for _, plugin := range m.pluginMap {
		go func(plugin Plugin) {
			defer func() {
				// 调用外部方法，必须做好recover工作
				if e := recover(); e != nil {
					log.Error("send modify contact panic: %+v", e)
				}
			}()
			plugin.modifyContact(contact)
		}(plugin)
	}
}

// broadcastMessage 广播message
func (m *Wechat) broadcastMessage(message datastruct.Message) {
	for _, plugin := range m.pluginMap {
		go func(plugin Plugin) {
			defer func() {
				// 调用外部方法，必须做好recover工作
				if e := recover(); e != nil {
					log.Error("send new message panic: %+v", e)
				}
			}()
			plugin.newMessage(message)
		}(plugin)
	}
}

func (m *Wechat) updateLoginStatus(item wwdk.LoginChannelItem) {
	// 做初步处理
	if item.Code == wwdk.LoginStatusErrorOccurred {
		// 登陆失败
		panic(fmt.Sprintf("WxWeb Login error: %+v", item.Err))
	}
	// 如果是登陆成功，则存一份联系人表
	if item.Code == wwdk.LoginStatusGotBatchContact {
		// 如果重新登陆了需要先清空原来的联系人，否则一定会造成联系人重复
		m.contacts = make(map[string]datastruct.Contact)
		m.syncContact(m.wechat.GetContactList())
	}
	// 更新到Wechat
	m.loginStatus = item
	// 广播loginStatus
	for _, plugin := range m.pluginMap {
		go func(plugin Plugin) {
			defer func() {
				// 调用外部方法，必须做好recover工作
				if e := recover(); e != nil {
					log.Error("send login status panic: %+v", e)
				}
			}()
			plugin.loginStatus(m.loginStatus)
		}(plugin)
	}
}

// 将给定的联系人处理后(主要操作是上传头像)同步到模块中
func (m *Wechat) syncContact(contacts []datastruct.Contact) {
	defer func() {
		if e := recover(); e != nil {
			log.Error("syncContact panic: %v", e)
		}
	}()
	contactChan := make(chan datastruct.Contact)
	for _, contact := range contacts {
		go func(contact datastruct.Contact) {
			sChan := make(chan string)
			go func() {
				fileurl, err := m.wechat.SaveContactImg(contact)
				if err == nil {
					sChan <- fileurl
				} else {
					log.Debug("wechat.SaveContactImg error: %v", err)
					// 如果出错，则输出一个空信息
					sChan <- ""
				}
			}()
			select {
			case ret := <-sChan:
				if ret != "" {
					contact.HeadImgURL = ret
				}
			case <-time.After(time.Second):
				// 超时，不做修改
			}
			contactChan <- contact
		}(contact)
	}
	for i := 0; i < len(contacts); i++ {
		c := <-contactChan
		m.contacts[c.UserName] = c
		m.broadcastContact(c)
	}
	log.Debug("syncContact form wwdk success, total(%d) contacts", len(contacts))
}

// checkToken 检查token是否有效
func (m *Wechat) checkToken(token string) (valid bool) {
	if _, ok := m.pluginMap[token]; ok {
		return true
	}
	return false
}

// 发送文字信息
// @Param toUserName 收件人userName
// @Param content 内容
// @return result 发送消息的返回，包括消息的本地与服务器id
// @return err 错误（为空则无错误
func (m *Wechat) sendTextMessage(token string, toUserName, content string) (result wechatstruct.SendMessageRespond, err string) {
	if !m.checkToken(token) {
		err = "token invalid"
		return
	}
	resp, e := m.wechat.SendTextMessage(toUserName, content)
	if e != nil {
		err = e.Error()
	} else {
		result = wechatstruct.SendMessageRespond{
			LocalID: resp.LocalID,
			MsgID:   resp.MsgID,
		}
	}
	return
}

// 发送文字信息
// @Param toUserName 收件人userName
// @Param content 内容
// @return result 撤回消息的返回，包含撤回消息的提示语句
// @return err 错误（为空则无错误
func (m *Wechat) revokeMessage(token string, srvMsgID, localMsgID, toUserName string) (result wechatstruct.RevokeMessageRespond, err string) {
	if !m.checkToken(token) {
		err = "token invalid"
		return
	}
	resp, e := m.wechat.SendRevokeMessage(srvMsgID, localMsgID, toUserName)
	if e != nil {
		err = e.Error()
	} else {
		result = wechatstruct.RevokeMessageRespond{
			Introduction: resp.Introduction,
			SysWording:   resp.SysWording,
		}
	}
	return
}

// 获取联系人列表
// @return result 联系人列表
// @return err 错误（为空则无错误
func (m *Wechat) getContactList(token string) (result []datastruct.Contact, err string) {
	if !m.checkToken(token) {
		err = "token invalid"
		return
	}
	for _, contact := range m.contacts {
		result = append(result, contact)
	}
	return
}

// 根据userName获取联系人
// @Param userName 要查找的联系人
// @return result 联系人列表
// @return err 错误（找不到联系人会返回User not found错误
func (m *Wechat) getContactByUserName(token string, userName string) (result datastruct.Contact, err string) {
	if !m.checkToken(token) {
		err = "token invalid"
		return
	}
	if contact, ok := m.contacts[userName]; ok {
		result = contact
	} else {
		err = "User not found"
	}
	return
}

// 根据alias获取联系人
// @Param alias 要查找的联系人
// @return result 联系人列表
// @return err 错误（找不到联系人会返回User not found错误
func (m *Wechat) getContactByAlias(token string, alias string) (result datastruct.Contact, err string) {
	if !m.checkToken(token) {
		err = "token invalid"
		return
	}
	found := false
	for _, v := range m.contacts {
		if v.Alias == alias {
			result = v
			found = true
		}
	}
	if !found {
		err = "User not found"
	}
	return
}

// 根据nickname获取联系人
// @Param nickname 要查找的联系人
// @return result 联系人列表
// @return err 错误（找不到联系人会返回User not found错误
func (m *Wechat) getContactByNickname(token string, nickname string) (result datastruct.Contact, err string) {
	if !m.checkToken(token) {
		err = "token invalid"
		return
	}
	found := false
	for _, v := range m.contacts {
		if v.NickName == nickname {
			result = v
			found = true
		}
	}
	if !found {
		err = "User not found"
	}
	return
}

// 根据remarkName获取联系人
// @Param remarkName 要查找的联系人
// @return result 联系人列表
// @return err 错误（找不到联系人会返回User not found错误
func (m *Wechat) getContactByRemarkName(token string, remarkName string) (result datastruct.Contact, err string) {
	if !m.checkToken(token) {
		err = "token invalid"
		return
	}
	found := false
	for _, v := range m.contacts {
		if v.RemarkName == remarkName {
			result = v
			found = true
		}
	}
	if !found {
		err = "User not found"
	}
	return
}

// 修改联系人昵称
// @Param userName 要修改的目标用户的userName
// @Param remarkName 要修改的昵称
// @return result none
// @return err 错误（为空则无错误
func (m *Wechat) modifyUserRemarkName(token string, userName, remarkName string) (result, err string) {
	if !m.checkToken(token) {
		err = "token invalid"
		return
	}
	_, e := m.wechat.ModifyUserRemakName(userName, remarkName)
	if e != nil {
		err = e.Error()
	}
	return
}

// 修改群标题
// @Param userName 要修改的目标用户的userName
// @Param newTopic 要修改的标题
// @return result none
// @return err 错误（为空则无错误
func (m *Wechat) modifyChatRoomTopic(token string, userName, newTopic string) (result, err string) {
	if !m.checkToken(token) {
		err = "token invalid"
		return
	}
	_, e := m.wechat.ModifyChatRoomTopic(userName, newTopic)
	if e != nil {
		err = e.Error()
	}
	return
}

// 获取运行信息
// @return result wwdk运行信息
// @return err 错误（为空则无错误
func (m *Wechat) getRunInfo(token string) (result wwdk.WechatRunInfo, err string) {
	if !m.checkToken(token) {
		err = "token invalid"
		return
	}
	result = m.wechat.GetRunInfo()
	return
}
