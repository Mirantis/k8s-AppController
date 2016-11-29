FROM alpine

RUN echo "@community http://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories
RUN apk --no-cache add runit@community &&\
    mkdir -p /etc/sv/ac &&\
    mkdir -p /opt/kubeac/manifests &&\
    touch /etc/sv/ac/down

ADD manifests /opt/kubeac/manifests
ADD ac_service.sh /etc/sv/ac/run
ADD run_runit.sh /usr/bin/run_runit
ADD ac-run.sh /usr/bin/ac-run
ADD ac-stop.sh /usr/bin/ac-stop
ADD kubeac /usr/bin/kubeac
