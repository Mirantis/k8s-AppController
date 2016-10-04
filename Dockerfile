FROM golang:alpine

RUN mkdir -p /go/src/github.com/Mirantis/k8s-AppController
COPY . /go/src/github.com/Mirantis/k8s-AppController

WORKDIR /go/src/github.com/Mirantis/k8s-AppController

RUN echo "@community http://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories

RUN apk --no-cache add git runit@community glide@community
RUN glide install --strip-vendor &&\
    go build -o kubeac &&\
    mv kubeac /usr/bin/kubeac &&\
    mkdir -p /opt/kubeac &&\
    mv /go/src/github.com/Mirantis/k8s-AppController/manifests /opt/kubeac/manifests &&\
    rm -fr /go


RUN mkdir -p /etc/sv/ac
ADD ac_service.sh /etc/sv/ac/run
ADD run_runit.sh /usr/bin/run_runit
ADD ac-run.sh /usr/bin/ac-run
ADD ac-stop.sh /usr/bin/ac-stop
RUN touch /etc/sv/ac/down
