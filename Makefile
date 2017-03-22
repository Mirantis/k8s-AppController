
IMAGE_REPO ?= mirantis/k8s-appcontroller
K8S_CLUSTER_MARKER = .k8s-cluster
K8S_VERSION ?= v1.5

.PHONY: docker
docker: kubeac Makefile
	docker build -t $(IMAGE_REPO) .

vendor: Makefile
	glide install --strip-vendor

test: vendor glide.lock Makefile
	go test ./cmd/...
	go test ./pkg/...

kubeac:
	bash hooks/pre_build

docker-publish:
	IMAGE_REPO=$(IMAGE_REPO) ./scripts/docker_publish.sh

.PHONY: img-in-dind
img-in-dind: docker $(K8S_CLUSTER_MARKER)
	IMAGE_REPO=$(IMAGE_REPO) ./scripts/import.sh

.PHONY: e2e
e2e: $(K8S_CLUSTER_MARKER) img-in-dind
	go test -c -o e2e.test ./e2e/
	./e2e.test --cluster-url=http://0.0.0.0:8080 --k8s-version=$(K8S_VERSION)

.PHONY: clean-all
clean-all: clean clean-k8s

.PHONY: clean
clean:
	-rm -f kubeac
	-rm e2e.test
	-docker rmi $(IMAGE_REPO)

.PHONY: clean-k8s
clean-k8s:
	./kubeadm-dind-cluster/fixed/dind-cluster-$(K8S_VERSION).sh clean
	-rm $(K8S_CLUSTER_MARKER)

$(K8S_CLUSTER_MARKER):
	if [ ! -d "kubeadm-dind-cluster" ]; then git clone https://github.com/Mirantis/kubeadm-dind-cluster.git; fi
	./kubeadm-dind-cluster/fixed/dind-cluster-$(K8S_VERSION).sh up
	touch $(K8S_CLUSTER_MARKER)
