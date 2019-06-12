module github.com/videocoin/cloud-accounts

go 1.12

require (
	github.com/codahale/hdrhistogram v0.0.0-20161010025455-3a0bb77429bd // indirect
	github.com/ethereum/go-ethereum v1.8.23
	github.com/gogo/protobuf v1.2.1
	github.com/golang/protobuf v1.3.1
	github.com/jinzhu/copier v0.0.0-20180308034124-7e38e58719c3
	github.com/jinzhu/gorm v1.9.8
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/opentracing/opentracing-go v1.1.0
	github.com/sirupsen/logrus v1.4.2
	github.com/streadway/amqp v0.0.0-20190404075320-75d898a42a94
	github.com/uber-go/atomic v1.4.0 // indirect
	github.com/videocoin/cloud-api v0.0.17
	github.com/videocoin/cloud-pkg v0.0.0-20190612175258-4c8b771d0bca
	github.com/videocoin/go-videocoin v0.0.0-20190612173315-a1e3cdbf3406
	google.golang.org/grpc v1.21.1
)

// replace github.com/VideoCoin/cloud-api => ../cloud-api
