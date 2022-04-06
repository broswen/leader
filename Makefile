.PHONY: build

REPO=broswen

test:
	go test ./...

docker-build:
	docker build -f Dockerfile.leader . -t $(REPO)/leader:latest
	docker build -f Dockerfile.worker . -t $(REPO)/worker:latest