FROM alpine

RUN echo "@community http://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories
RUN apk --no-cache add runit@community &&\
    mkdir -p /etc/sv/ac &&\
    mkdir -p /opt/kubeac/manifests &&\
    touch /etc/sv/ac/down

ADD manifests /opt/kubeac/manifests
ADD kubeac /usr/bin/kubeac
