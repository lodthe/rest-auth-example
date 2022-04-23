.PHONY: proto build

all: build docker-publish-x86

build:
	go build -o bin/server cmd/server/*
	go build -o bin/worker cmd/worker/*

docker-publish-x86:
	docker buildx build --platform linux/x86_64 -t lodthe/rest-auth-example-server -f dockerfiles/Dockerfile-server .
	docker push lodthe/rest-auth-example-server
	@echo "Server image published"

	docker buildx build --platform linux/x86_64 -t lodthe/rest-auth-example-worker -f dockerfiles/Dockerfile-worker .
	docker push lodthe/rest-auth-example-worker
	@echo "Worker image published"
