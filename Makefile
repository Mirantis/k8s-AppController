
TAG ?= mirantis/k8s-appcontroller
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
	cd e2e && go test

.PHONY: test-all
test-all: test e2e

.PHONY: clean-all
clean-all: clean clean-k8s

.PHONY: clean
clean:
	rm -f kubeac
	docker rmi $(TAG)


.PHONY: clean-k8s
clean-k8s: k8s-down
	<$(K8S_SOURCE_LOCATION) xargs rm -rf
	rm $(K8S_SOURCE_LOCATION)


$(K8S_SOURCE_LOCATION):
	scripts/checkout_k8s.sh $(K8S_SOURCE_LOCATION)

$(K8S_CLUSTER_MARKER): $(K8S_SOURCE_LOCATION)
	<$(K8S_SOURCE_LOCATION) xargs -I % bash -c "cd %/kubernetes && dind/dind-cluster.sh up"

.PHONY: k8s-down
k8s-down:
	<$(K8S_SOURCE_LOCATION) xargs -I % bash -c "cd %/kubernetes && dind/dind-cluster.sh down"
