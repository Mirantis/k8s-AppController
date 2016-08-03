FROM golang

RUN mkdir -p /go/src/appcontroller
COPY . /go/src/appcontroller

WORKDIR /go/src/appcontroller

RUN go get ./...
RUN go build -o kubeac cmd/kubeac/main.go
RUN go build -o wrap cmd/wrap/main.go
RUN ln -s kubeac /usr/bin/kubeac
RUN ln -s wrap /usr/bin/wrap
