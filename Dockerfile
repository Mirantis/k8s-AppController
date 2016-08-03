FROM golang

RUN mkdir -p /go/src/appcontroller
COPY . /go/src/appcontroller

WORKDIR /go/src/appcontroller

RUN go get ./...
RUN go build cmd/kubeac/main.go
RUN ln -s main /usr/bin/kubeac
