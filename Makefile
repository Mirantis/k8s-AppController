.PHONY: docker

TAG ?= mirantis/k8s-appcontroller

docker: vendor Makefile
	docker build -t $(TAG) .

vendor: Makefile
	glide install --strip-vendor

test: vendor glide.lock Makefile
	go list ./... | grep -v /vendor/ | xargs go test
