FROM golang:alpine

RUN mkdir -p /go/src/appcontroller
COPY . /go/src/appcontroller

WORKDIR /go/src/appcontroller

RUN apk --no-cache add git
RUN go get ./...
RUN go build -o kubeac cmd/kubeac/main.go
RUN go build -o wrap cmd/wrap/main.go
RUN ln -s kubeac /usr/bin/kubeac
RUN ln -s wrap /usr/bin/wrap

RUN echo "@testing http://dl-4.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories
RUN apk --no-cache add runit@testing

RUN mkdir -p /etc/sv/ac
ADD ac_service.sh /etc/sv/ac/run
ADD ac-run.sh /usr/bin/ac-run
ADD ac-stop.sh /usr/bin/ac-stop
RUN touch /etc/sv/ac/down
