
.PHONY: build
build:
	mkdir -p dist/
	go build -o dist/redis-proxy

.PHONY: test
test:
	go test github.com/eastside-eng/redis-proxy/cache
	go test github.com/eastside-eng/redis-proxy/proxy
	docker-compose up -d
	# Sleep 15 seconds to let the servers come up...
	sleep 15
	# End to end tests are in the main package
	go test github.com/eastside-eng/redis-proxy

.PHONY: run-dev
run-dev:
	docker-compose up --build

.PHONY: run
run:
	docker-compose up

.PHONY: cli
cli:
	docker run -it --link redis:redis --rm redis redis-cli -h redis -p 6379
