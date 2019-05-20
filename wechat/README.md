# wechat模块说明

wechat模块基于[wwdk](https://github.com/ikuiki/wwdk)包与微信网页版服务器交互，维持登陆态，收发新消息等，模块本身没有处理消息的业务逻辑，需要Plugin接入后在Plugin中实现业务逻辑；对于微信中的媒体，因为必须要带登陆态（cookie）的http client才能下载，所以就设计为在服务器下载，然后传给uploader，uploader上传完成后返回资源url后再将信息传给Plugin（可能会产生延迟）

## Plugin

Plugin为实现微信业务逻辑的主要构成，Plugin需要在wechat模块注册后才能开始使用微信功能。Plugin根据连接方式，可分为rpcPlugin与mqttPlugin，rpcPlugin是基于mqant框架中Module的，调用wechat模块的方式遵循mqant中的定义。mqtt模块是基于mqant框架中的gate模块的mqtt客户端的，调用wechat模块的方式遵循mqtt中客户端与服务器的调用定义，并且相对于rpcPlugin，mqttPlugin在连接后需要先登陆session再注册到wechat模块。rpcPlugin与mqttPlugin注册到wechat模块后都会获得wechatToken，后续所有对wechat的调用操作都需要使用这个token

### rpcPlugin

#### RegisterRpcPlugin

注册插件到wechat

request:
| Param                    | Type   | Description                                                                             |
| ------------------------ | ------ | --------------------------------------------------------------------------------------- |
| name                     | string | 插件的名称                                                                              |
| description              | string | 插件的描述                                                                              |
| moduleType               | string | 插件的mqant module类型                                                                  |
| loginListenerFunc        | string | 监听登陆状态的方法名称，如果有新的登陆状态会以登陆状态为参数调用该方法                  |
| contactListenerFunc      | string | 联系人变化监听方法，如果有联系人变化会以联系人为参数调用该方法                          |
| msgListenerFunc          | string | 新信息监听方法，如果有新消息到达会以新消息为参数调用该方法                              |
| addPluginListenerFunc    | string | 新插件注册监听方法，如果有新wechat Plugin注册会以该插件的信息为参数调用该方法           |
| removePluginListenerFunc | string | 现有插件移除监听方法，如果有已注册的wechat Plugin移除，会以该插件的信息为参数调用该方法 |

response:
| Return Value | Type   | Description                             |
| ------------ | ------ | --------------------------------------- |
| token        | string | wechatToken，后续调用微信方法需要用到的 |
| err          | string | 错误消息（如果没有则为空）              |

其中loginListenerFunc期望注册的是一个以wwdk.LoginChannelItem为参数的方法
其中contactListenerFunc期望注册的是一个以wwdk/datastruct.Contact为参数的方法
其中msgListenerFunc期望注册的是一个以wwdk/datastruct.Message为参数的方法
其中addPluginListenerFunc期望注册的是一个以PluginDesc为参数的方法
其中removePluginListenerFunc期望注册的是一个以PluginDesc为参数的方法

#### Plugin_GetPluginList

获取已注册的插件

request:
| Param | Type   | Description                 |
| ----- | ------ | --------------------------- |
| token | string | wechatToken，注册时获取到的 |

response:
| Return Value | Type         | Description                |
| ------------ | ------------ | -------------------------- |
| list         | []PluginDesc | 插件描述（数组）           |
| err          | string       | 错误消息（如果没有则为空） |

#### Wechat_SendTextMessage

发送文字信息

request:
| Param      | Type   | Description                 |
| ---------- | ------ | --------------------------- |
| token      | string | wechatToken，注册时获取到的 |
| toUserName | string | 目标用户的微信userName      |
| content    | string | 内容                        |

response:
| Return Value | Type                            | Description                             |
| ------------ | ------------------------------- | --------------------------------------- |
| result       | wechatstruct.SendMessageRespond | 发送信息后的返回，内有微信的messageID等 |
| err          | string                          | 错误消息（如果没有则为空）              |

#### Wechat_RevokeMessage

撤回消息

request:
| Param | Type   | Description                 |
| ----- | ------ | --------------------------- |
| token | string | wechatToken，注册时获取到的 |
| srvMsgID   | string | 要撤回的消息的服务器ID      |
| localMsgID | string | 要撤回的消息的本地ID        |
| toUserName | string | 收件人userName              |

response:
| Return Value | Type                              | Description                            |
| ------------ | --------------------------------- | -------------------------------------- |
| result       | wechatstruct.RevokeMessageRespond | 撤回消息的返回，包含撤回消息的提示语句 |
| err          | string                            | 错误（为空则无错误                     |

#### Wechat_GetUser

获取登陆用户

request:
| Param      | Type   | Description                 |
| ---------- | ------ | --------------------------- |
| token      | string | wechatToken，注册时获取到的 |

response:
| Return Value | Type            | Description        |
| ------------ | --------------- | ------------------ |
| result       | datastruct.User | 用户信息           |
| err          | string          | 错误（为空则无错误 |

#### Wechat_GetContactList

获取联系人列表

request:
| Param | Type   | Description                 |
| ----- | ------ | --------------------------- |
| token | string | wechatToken，注册时获取到的 |

response:
| Return Value | Type                 | Description        |
| ------------ | -------------------- | ------------------ |
| result       | []datastruct.Contact | 联系人列表         |
| err          | string               | 错误（为空则无错误 |

#### Wechat_GetContactByUserName

通过UserName获取联系人

request:
| Param    | Type   | Description                 |
| -------- | ------ | --------------------------- |
| token    | string | wechatToken，注册时获取到的 |
| userName | string | 要查询的UserName            |

response:
| Return Value | Type               | Description        |
| ------------ | ------------------ | ------------------ |
| result       | datastruct.Contact | 目标联系人         |
| err          | string             | 错误（为空则无错误 |

#### Wechat_GetContactByAlias

通过Alias获取联系人

request:
| Param | Type   | Description                 |
| ----- | ------ | --------------------------- |
| token | string | wechatToken，注册时获取到的 |
| alias | string | 要查询的Alias               |

response:
| Return Value | Type               | Description        |
| ------------ | ------------------ | ------------------ |
| result       | datastruct.Contact | 目标联系人         |
| err          | string             | 错误（为空则无错误 |

#### Wechat_GetContactByNickname

通过Nickname获取联系人

request:
| Param    | Type   | Description                 |
| -------- | ------ | --------------------------- |
| token    | string | wechatToken，注册时获取到的 |
| nickname | string | 要查询的Nickname            |

response:
| Return Value | Type               | Description        |
| ------------ | ------------------ | ------------------ |
| result       | datastruct.Contact | 目标联系人         |
| err          | string             | 错误（为空则无错误 |


#### Wechat_GetContactByRemarkName

通过RemarkName获取联系人

request:
| Param      | Type   | Description                 |
| ---------- | ------ | --------------------------- |
| token      | string | wechatToken，注册时获取到的 |
| remarkName | string | 要查询的RemarkName          |

response:
| Return Value | Type               | Description        |
| ------------ | ------------------ | ------------------ |
| result       | datastruct.Contact | 目标联系人         |
| err          | string             | 错误（为空则无错误 |

#### Wechat_ModifyUserRemarkName

修改指定联系人的RemarkName

request:
| Param      | Type   | Description                 |
| ---------- | ------ | --------------------------- |
| token      | string | wechatToken，注册时获取到的 |
| userName   | string | 要修改的目标用户的userName  |
| remarkName | string | 要修改的昵称                |

response:
| Return Value | Type   | Description          |
| ------------ | ------ | -------------------- |
| result       | string | 无内容，仅为了占位用 |
| err          | string | 错误（为空则无错误   |

#### Wechat_ModifyChatRoomTopic

修改群标题

request:
| Param      | Type   | Description                 |
| ---------- | ------ | --------------------------- |
| token      | string | wechatToken，注册时获取到的 |
| userName   | string | 要修改的目标群的userName    |
| remarkName | string | 要修改的标题                |

response:
| Return Value | Type   | Description          |
| ------------ | ------ | -------------------- |
| result       | string | 无内容，仅为了占位用 |
| err          | string | 错误（为空则无错误   |

#### Wechat_GetRunInfo

获取wwdk的运行信息

request:
| Param | Type   | Description                 |
| ----- | ------ | --------------------------- |
| token | string | wechatToken，注册时获取到的 |

response:
| Return Value | Type               | Description                      |
| ------------ | ------------------ | -------------------------------- |
| result       | wwdk.WechatRunInfo | wwdk的运行信息，具体请参考wwdk包 |
| err          | string             | 错误（为空则无错误               |

### mqttPlugin


## Uploader
