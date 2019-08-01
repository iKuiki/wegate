module wegate

require (
	github.com/eclipse/paho.mqtt.golang v1.2.0
	github.com/getsentry/sentry-go v0.2.1
	github.com/google/uuid v1.1.1
	github.com/ikuiki/go-component v0.0.0-20171218165758-b9f2562e71d1
	github.com/ikuiki/storer v1.0.0
	github.com/ikuiki/wwdk v2.6.5+incompatible
	github.com/liangdas/armyant v0.0.0-20181120080818-50ccc5936868
	github.com/liangdas/mqant v1.8.1
	github.com/mdp/qrterminal v1.0.1
	github.com/pkg/errors v0.8.1
	github.com/pquerna/otp v1.2.0
	golang.org/x/crypto v0.0.0-20190308221718-c2843e01d9a2
	golang.org/x/net v0.0.0-20190424112056-4829fb13d2c6 // indirect
)

replace github.com/liangdas/mqant => github.com/ikuiki/mqant v1.8.1-0.20190427142930-7dabfa32d064
