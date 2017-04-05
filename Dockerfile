FROM alpine

RUN mkdir -p /opt/kubeac/manifests

ADD manifests /opt/kubeac/manifests
ADD kubeac /usr/bin/kubeac
