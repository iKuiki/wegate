# wegate 微信web网关

[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=flat)](LICENSE)

---

基于[wwdk](https://github.com/ikuiki/wwdk)包与[mqant v1.8.0 小幅改版](https://github.com/liangdas/mqant)包做的微信web网关，初衷是因为直接在app中使用wwdk包登陆，如果app包逻辑过于复杂有bug导致崩溃，重启后又因别的原因未能登陆，重新登陆后就会导致userName变化，所以求稳为主的话，将业务逻辑和微信web客户端隔离为2个项目就很有必要了，然后业务逻辑再以远程调用的形式接入到此项目中获取微信web的访问，就能做到更佳的稳定性

为了方便接入，使用了mqant框架来作为接入网关，使用mqtt协议很适合微信网关这个场景

---

在mqant的基础上，我另外设计了2个结构：微信插件（wegate/wechat.Plugin）与微信上传器（wegate/wechat.Uploader)，微信插件用于封装微信处理逻辑，uploader用于上传微信中遇到的媒体

## Plugin

微信的Plugin分为本地Plugin（rpcPlugin）与远程Plugin（mqttPlugin），本地Plugin是运行在mqant的module中，而远程Plugin则是运行在另外的项目中通过mqtt协议与wegate连接

如果是mqtt客户端，连接到wegate后必须先登陆(Login模块)

不论是rpc客户端还是mqtt客户端，连接到wegate以后都需要注册到wechat模块获取token后才可以调用wechat功能
