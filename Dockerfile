FROM golang:1.12.4 as builder
WORKDIR /go/src/github.com/videocoin/cloud-accounts
COPY . .
RUN make build

FROM bitnami/minideb:jessie
RUN apt-get update && apt-get -y install ca-certificates
COPY --from=builder /go/src/github.com/videocoin/cloud-accounts/bin/accounts /opt/videocoin/bin/accounts
CMD ["/opt/videocoin/bin/accounts"]