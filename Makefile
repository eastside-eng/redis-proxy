
.PHONY: build
build:
	mkdir -p dist/
	go build -o dist/redis-proxy

.PHONY: test
test:
	go test github.com/eastside-eng/redis-proxy/cache
	go test github.com/eastside-eng/redis-proxy/proxy
	docker-compose up

.PHONY: run-dev
run-dev:
	docker-compose up --build

.PHONY: run
run:
	docker-compose up

.PHONY: cli
cli:
	docker run -it --link redis:redis --rm redis redis-cli -h redis -p 6379
