OS=linux
ARCH=amd64
BUILDKIT_PROGRESS=plain
IMG_VER?=latest

image-prod:
	docker build -t theshamuel/file-hash-agent:${IMG_VER} .

image-dev:
	docker build -t theshamuel/file-hash-agent:${IMG_VER} --build-arg SKIP_TESTS=true .

deploy:
	docker-compose up -d

lint:
	$(GOPATH)/bin/golangci-lint run --config .golangci.yml

unit-test:
	go test -mod=vendor --coverprofile cover.out ./...

race-test:
	&& go test -race -mod=vendor -timeout=60s -count 1 ./...

.PHONY: image-dev deploy