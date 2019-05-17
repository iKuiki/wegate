module wegate

require (
	github.com/eclipse/paho.mqtt.golang v1.2.0
	github.com/ikuiki/go-component v0.0.0-20171218165758-b9f2562e71d1
	github.com/ikuiki/wwdk v2.3.3+incompatible
	github.com/liangdas/armyant v0.0.0-20181120080818-50ccc5936868
	github.com/liangdas/mqant v1.8.1
	github.com/mdp/qrterminal v1.0.1
	github.com/pkg/errors v0.8.1
	golang.org/x/net v0.0.0-20190424112056-4829fb13d2c6 // indirect
	golang.org/x/sys v0.0.0-20190426135247-a129542de9ae // indirect
)

// 解决国内无法下载的几个包
replace (
	golang.org/x/crypto => github.com/golang/crypto v0.0.0-20190513172903-22d7a77e9e5f
	golang.org/x/net => github.com/golang/net v0.0.0-20190514140710-3ec191127204
	golang.org/x/sync => github.com/golang/sync v0.0.0-20190423024810-112230192c58
	golang.org/x/sys => github.com/golang/sys v0.0.0-20190514135907-3a4b5fb9f71f
	golang.org/x/text => github.com/golang/text v0.3.2
	golang.org/x/tools => github.com/golang/tools v0.0.0-20190515035509-2196cb7019cc
	google.golang.org/appengine => github.com/golang/appengine v1.6.0
)

replace github.com/liangdas/mqant => github.com/ikuiki/mqant v1.8.1-0.20190427142930-7dabfa32d064
