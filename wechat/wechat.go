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
	loginChan := make(chan wwdk.LoginChannelItem, 10)
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
	syncChannel := make(chan wwdk.SyncChannelItem, 100)
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
					if len(contact.MemberList) > 0 {
						m.saveMembersImg(&contact)
					}
					m.userInfo.contactList[contact.UserName] = contact
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
	// 复制一份pluginMap防止发生同时读写map的异常
	pMap := m.pluginMap
	for _, plugin := range pMap {
		go func(plugin Plugin) {
			defer func() {
				// 调用外部方法，必须做好recover工作
				if e := recover(); e != nil {
					log.Error("send modify contact panic: %+v", e)
				}
			}()
			plugin.modifyContactEvent(contact)
		}(plugin)
	}
}

// broadcastMessage 广播message
func (m *Wechat) broadcastMessage(message datastruct.Message) {
	// 复制一份pluginMap防止发生同时读写map的异常
	pMap := m.pluginMap
	for _, plugin := range pMap {
		go func(plugin Plugin) {
			defer func() {
				// 调用外部方法，必须做好recover工作
				if e := recover(); e != nil {
					log.Error("send new message panic: %+v", e)
				}
			}()
			plugin.newMessageEvent(message)
		}(plugin)
	}
}

func (m *Wechat) updateLoginStatus(item wwdk.LoginChannelItem) {
	// 对code进行预处理
	switch item.Code {
	case wwdk.LoginStatusErrorOccurred:
		// 登陆失败
		panic(fmt.Sprintf("WxWeb Login error: %+v", item.Err))
	case wwdk.LoginStatusWaitForScan:
		// 如果重新登陆了，就重置用户信息
		m.userInfo.user = nil
	case wwdk.LoginStatusInitFinish:
		// 获取用户信息
		u, e := m.wechat.GetUser()
		if e != nil {
			log.Error("get user fail: %v", e)
		} else {
			m.userInfo.user = &u
			// 同步用户头像
			m.syncUser()
		}
	case wwdk.LoginStatusGotContact:
		// 如果是通过存储信息免登录，则不会有Init Finish的状态，则再此记录用户信息
		if m.userInfo.user == nil {
			// 获取用户信息
			u, e := m.wechat.GetUser()
			if e != nil {
				log.Error("get user fail: %v", e)
			} else {
				m.userInfo.user = &u
				// 同步用户头像
				m.syncUser()
			}
		}
	case wwdk.LoginStatusGotBatchContact:
		// 如果是登陆成功，则存一份联系人表
		// 如果重新登陆了需要先清空原来的联系人，否则一定会造成联系人重复
		contacts := m.wechat.GetContactList()
		m.userInfo.contactList = make(map[string]datastruct.Contact)
		// 先保存一份没有头像的，可以给先到达的消息用
		for _, contact := range contacts {
			m.userInfo.contactList[contact.UserName] = contact
		}
		m.controlSigChan <- controlSigUploadContactImg // 发控制信号要求上传头像
	}
	// 更新到Wechat
	m.loginStatus = item
	// 广播loginStatus
	// 复制一份pluginMap防止发生同时读写map的异常
	pMap := m.pluginMap
	for _, plugin := range pMap {
		go func(plugin Plugin) {
			defer func() {
				// 调用外部方法，必须做好recover工作
				if e := recover(); e != nil {
					log.Error("send login status panic: %+v", e)
				}
			}()
			plugin.loginStatusEvent(m.loginStatus)
		}(plugin)
	}
}

// 处理当前登陆用户(主要操作是上传头像)
func (m *Wechat) syncUser() {
	if m.userInfo.user == nil {
		return
	}
	sChan := make(chan string)
	go func() {
		fileurl, err := m.wechat.SaveUserImg(*m.userInfo.user)
		if err == nil {
			sChan <- fileurl
		} else {
			sChan <- ""
		}
	}()
	select {
	case ret := <-sChan:
		if ret != "" {
			m.userInfo.user.HeadImgURL = ret
		}
	case <-time.After(time.Minute):
		// 超时，不做修改
	}
}

func (m *Wechat) saveContactImg(contact *datastruct.Contact) {
	sChan := make(chan string)
	go func() {
		fileurl, err := m.wechat.SaveContactImg(*contact)
		if err == nil {
			sChan <- fileurl
		} else {
			if err.Error() != "mediaStorer.Storer error: uploader not found" {
				log.Error("wechat.SaveContactImg error: %v", err)
			}
			// 如果出错，则输出一个空信息
			sChan <- ""
		}
	}()
	select {
	case ret := <-sChan:
		if ret != "" {
			contact.HeadImgURL = ret
		}
	case <-time.After(time.Minute): // 此处超时可以忽略，因为SaveContact时已经做了超时判断
		// 超时，不做修改
		log.Debug("save Contact %s HeadImg timeout, skip...", contact.NickName)
	}
}

func (m *Wechat) saveMembersImg(contact *datastruct.Contact) {
	// 清空后再赋值
	members := contact.MemberList
	contact.MemberList = []datastruct.Member{}

	preMemberChan := make(chan datastruct.Member)
	memberChan := make(chan datastruct.Member)
	for i := 0; i < 5; i++ { // 5并行避免抢占带宽造成大面积timeout
		go func(workNo int) {
			for member := range preMemberChan {
				sChan := make(chan string)
				go func() {
					fileurl, err := m.wechat.SaveMemberImg(member, contact.UserName)
					if err == nil {
						sChan <- fileurl
					} else {
						if err.Error() != "mediaStorer.Storer error: uploader not found" {
							log.Error("wechat.SaveMemberImg error: %v", err)
						}
						// 如果出错，则输出一个空信息
						sChan <- ""
					}
				}()
				select {
				case ret := <-sChan:
					if ret != "" {
						member.KeyWord = ret
					}
				case <-time.After(time.Minute): // 此处超时可以忽略，因为SaveContact时已经做了超时判断
					// 超时，不做修改
					log.Debug("save Member %s HeadImg timeout, skip...", member.NickName)
				}
				memberChan <- member
			}
		}(i)
	}
	go func() {
		for _, member := range members {
			preMemberChan <- member
		}
		close(preMemberChan)
	}()
	for i := 0; i < len(members); i++ {
		mb := <-memberChan
		contact.MemberList = append(contact.MemberList, mb)
	}
	close(memberChan)
}

// 将给定的联系人处理后(主要操作是上传头像)同步到模块中
func (m *Wechat) syncContact(contacts []datastruct.Contact) {
	defer func() {
		if e := recover(); e != nil {
			log.Error("syncContact panic: %v", e)
		}
	}()
	preContactChan := make(chan datastruct.Contact)
	contactChan := make(chan datastruct.Contact)
	for i := 0; i < 5; i++ { // 5并行下载
		go func(workNo int) {
			// preContactChan
			for contact := range preContactChan {
				m.saveContactImg(&contact)
				contactChan <- contact
			}
		}(i)
	}
	go func() {
		for _, contact := range contacts {
			preContactChan <- contact
		}
		close(preContactChan)
	}()
	for i := 0; i < len(contacts); i++ {
		c := <-contactChan
		if len(c.MemberList) > 0 {
			m.saveMembersImg(&c)
		}
		m.userInfo.contactList[c.UserName] = c
		m.broadcastContact(c)
	}
	close(contactChan)
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
// @Param srvMsgID 要撤回的消息的服务器ID
// @Param localMsgID 要撤回的消息的本地ID
// @Param toUserName 收件人userName
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

// 获取当前登陆用户
// @return result 当前登陆用户
// @return err 错误（为空则无错误
func (m *Wechat) getUser(token string) (result datastruct.User, err string) {
	if !m.checkToken(token) {
		err = "token invalid"
		return
	}
	if m.userInfo.user == nil {
		err = "User not found"
		return
	}
	result = *m.userInfo.user
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
	for _, contact := range m.userInfo.contactList {
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
	if contact, ok := m.userInfo.contactList[userName]; ok {
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
	for _, v := range m.userInfo.contactList {
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
	for _, v := range m.userInfo.contactList {
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
	for _, v := range m.userInfo.contactList {
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
// @Param userName 要修改的目标群的userName
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
