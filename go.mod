module github.com/videocoin/cloud-accounts

go 1.12

require (
	cloud.google.com/go v0.37.4 // indirect
	github.com/ethereum/go-ethereum v1.8.27
	github.com/golang/protobuf v1.3.1
	github.com/jinzhu/copier v0.0.0-20180308034124-7e38e58719c3
	github.com/jinzhu/gorm v1.9.12
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/opentracing/opentracing-go v1.1.0
	github.com/sirupsen/logrus v1.4.2
	github.com/streadway/amqp v0.0.0-20190404075320-75d898a42a94
	github.com/videocoin/cloud-api v0.2.15
	github.com/videocoin/cloud-pkg v0.0.6
	github.com/videocoin/go-faucet v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.21.1
	gopkg.in/yaml.v2 v2.2.2 // indirect
)

replace github.com/videocoin/cloud-api => ../cloud-api

replace github.com/videocoin/cloud-pkg => ../cloud-pkg

replace github.com/videocoin/go-faucet => ../go-faucet
