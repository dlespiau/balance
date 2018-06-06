all: build/proxy build/service docker

# go build is clever enough (esp. since 1.10) to not build too much, just force
# make to always call go build.
build/%: FORCE
	mkdir -p build
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -i -o build/$* ./cmd/$*

FORCE:

docker: .docker-proxy.done .docker-service.done

.docker-%.done: docker/Dockerfile.%
	cp $^ build/
	docker build -t quay.io/damien.lespiau/balance/$* -f build/Dockerfile.$* ./build

publish: all
	docker push quay.io/damien.lespiau/balance/proxy
	docker push quay.io/damien.lespiau/balance/service

clean: clean-build clean-docker
clean-build:
	rm -rf build
clean-docker:
	rm -rf .docker-*.done

.PHONY: all publish clean clean-build clean-docker
