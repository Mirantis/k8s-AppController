FROM golang:alpine

RUN mkdir -p /go/src/github.com/Mirantis/k8s-AppController
COPY . /go/src/github.com/Mirantis/k8s-AppController

WORKDIR /go/src/github.com/Mirantis/k8s-AppController

RUN apk --no-cache add git
RUN go get ./... && go build -o kubeac cmd/kubeac/main.go && go build -o wrap cmd/wrap/main.go && go build -o bootstrap cmd/bootstrap/main.go && go build -o get-status cmd/get-status/main.go && mv kubeac /usr/bin/kubeac && mv wrap /usr/bin/wrap && mv bootstrap /usr/bin/bootstrap && mv get-status /usr/bin/get-status && mkdir -p /opt/kubeac && mv /go/src/github.com/Mirantis/k8s-AppController/manifests /opt/kubeac/manifests && rm -fr /go

RUN echo "@community http://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories
RUN apk --no-cache add runit@community

RUN mkdir -p /etc/sv/ac
ADD ac_service.sh /etc/sv/ac/run
ADD run_runit.sh /usr/bin/run_runit
ADD ac-run.sh /usr/bin/ac-run
ADD ac-stop.sh /usr/bin/ac-stop
RUN touch /etc/sv/ac/down
