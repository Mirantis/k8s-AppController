FROM golang:alpine

RUN mkdir -p /go/src/github.com/Mirantis/k8s-AppController
COPY . /go/src/github.com/Mirantis/k8s-AppController

WORKDIR /go/src/github.com/Mirantis/k8s-AppController

RUN apk --no-cache add git
RUN go get ./...
RUN go build -o kubeac cmd/kubeac/main.go
RUN go build -o wrap cmd/wrap/main.go
RUN mv kubeac /usr/bin/kubeac
RUN mv wrap /usr/bin/wrap

RUN echo "@community http://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories
RUN apk --no-cache add runit@community

RUN mkdir -p /etc/sv/ac
ADD ac_service.sh /etc/sv/ac/run
ADD run_runit.sh /usr/bin/run_runit
ADD ac-run.sh /usr/bin/ac-run
ADD ac-stop.sh /usr/bin/ac-stop
RUN touch /etc/sv/ac/down
