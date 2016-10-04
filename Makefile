.PHONY: docker

TAG ?= mirantis/k8s-appcontroller

docker: vendor glide.lock Makefile
	docker build -t $(TAG) .

vendor: Makefile
	glide install --strip-vendor
	glide-vc --only-code --no-tests
