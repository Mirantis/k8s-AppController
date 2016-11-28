FROM alpine

RUN echo "@community http://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories
RUN apk --no-cache add runit@community

RUN mkdir -p /etc/sv/ac
RUN mkdir -p /opt/kubeac
ADD manifests /opt/kubeac
ADD ac_service.sh /etc/sv/ac/run
ADD run_runit.sh /usr/bin/run_runit
ADD ac-run.sh /usr/bin/ac-run
ADD ac-stop.sh /usr/bin/ac-stop
RUN touch /etc/sv/ac/down
ADD kubeac /bin/kubeac

