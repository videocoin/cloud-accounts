module github.com/videocoin/cloud-accounts

go 1.12

require (
	github.com/cespare/cp v1.1.1 // indirect
	github.com/edsrzf/mmap-go v1.0.0 // indirect
	github.com/ethereum/go-ethereum v1.8.27
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.3.1
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0
	github.com/jinzhu/copier v0.0.0-20180308034124-7e38e58719c3
	github.com/jinzhu/gorm v1.9.9
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/opentracing/opentracing-go v1.1.0
	github.com/sirupsen/logrus v1.4.2
	github.com/streadway/amqp v0.0.0-20190404075320-75d898a42a94
	github.com/videocoin/cloud-api v0.2.15
	github.com/videocoin/cloud-pkg v0.0.6
	google.golang.org/grpc v1.21.1
)

replace github.com/videocoin/cloud-api => ../cloud-api

replace github.com/videocoin/cloud-pkg => ../cloud-pkg
