
.PHONY: build
build:
	mkdir -p dist/
	go build -o dist/redis-proxy

.PHONY: test
test:
	go test github.com/eastside-eng/redis-proxy/cache
	go test github.com/eastside-eng/redis-proxy/proxy

run:
	docker run -p 6379:6379 -d --name redis redis || docker start redis
	./dist/redis-proxy

cli:
	docker run -it --link redis:redis --rm redis redis-cli -h redis -p 6379
