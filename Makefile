
TAG ?= mirantis/k8s-appcontroller
WORKING ?= ~/testappcontroller
K8S_SOURCE_LOCATION = .k8s-source
K8S_CLUSTER_MARKER = .k8s-cluster

.PHONY: docker
docker: kubeac Makefile
	docker build -t $(TAG) .

vendor: Makefile
	glide install --strip-vendor

test: vendor glide.lock Makefile
	go test ./cmd/...
	go test ./pkg/...

kubeac:
	bash hooks/pre_build

.PHONY: img-in-dind
img-in-dind: docker $(K8S_CLUSTER_MARKER)
	IMAGE_REPO=mirantis/k8s-appcontroller bash scripts/import.sh

.PHONY: e2e
e2e: $(K8S_CLUSTER_MARKER) img-in-dind
	go test -c -o e2e.test ./e2e/
	PATH=$(PATH):$(WORKING)/kubernetes/_output/bin/ ./e2e.test --cluster-url=http://0.0.0.0:8888

.PHONY: clean-all
clean-all: clean clean-k8s

.PHONY: clean
clean:
	rm -f kubeac
	-docker rmi $(TAG)

.PHONY: clean-k8s
clean-k8s:
	<$(K8S_SOURCE_LOCATION) xargs rm -rf
	rm $(K8S_SOURCE_LOCATION)
	rm $(K8S_CLUSTER_MARKER)

$(K8S_SOURCE_LOCATION):
	WORKING=$(WORKING) scripts/checkout_k8s.sh > $(K8S_SOURCE_LOCATION)

$(K8S_CLUSTER_MARKER): $(K8S_SOURCE_LOCATION)
	WORKING=$(WORKING) ./scripts/prepare_dind.sh
	touch $(K8S_CLUSTER_MARKER)
