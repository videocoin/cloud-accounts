GOOS?=linux
GOARCH?=amd64

export CGO_ENABLED=1

NAME=vc_accounts
VERSION=$$(git describe --abbrev=0)-$$(git rev-parse --short HEAD)

DBM_MSQLURI=host=127.0.0.1 user=mysql dbname=videocoin sslmode=disable password=mysql

version:
	@echo ${VERSION}

build:
	GOOS=${GOOS} GOARCH=${GOARCH} \
		go build \
			-ldflags="-w -s -X main.Version=${VERSION}" \
			-o bin/${NAME} \
			./cmd/main.go

build-dev:
	env GO111MODULE=on GOOS=${GOOS} GOARCH=${GOARCH} \
		go build \
			-ldflags="-w -s -X main.Version=${VERSION}" \
			-o bin/${NAME} \
			./cmd/main.go

deps:
	env GO111MODULE=on go mod vendor
	# https://github.com/ethereum/go-ethereum/issues/2738
	cp -r $(GOPATH)/src/github.com/VideoCoin/go-videocoin/crypto/secp256k1/libsecp256k1 \
	vendor/github.com/VideoCoin/go-videocoin/crypto/secp256k1/

docker-build:
	docker build -t gcr.io/${GCP_PROJECT}/${NAME}:${VERSION} -f Dockerfile .

docker-push:
	gcloud docker -- push gcr.io/${GCP_PROJECT}/${NAME}:${VERSION}

dbm-status:
	goose -dir migrations -table ${NAME} postgres "${DBM_MSQLURI}" status

dbm-up:
	goose -dir migrations -table ${NAME} postgres "${DBM_MSQLURI}" up

dbm-down:
	goose -dir migrations -table ${NAME} postgres "${DBM_MSQLURI}" down

release: build docker-build docker-push
