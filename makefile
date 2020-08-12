SHELL := /bin/bash

build:
	[[ -d .build ]] || mkdir .build
	-rm ./.build/app
	CGO_ENABLED=0 GOOS=linux go build -o ./app-service ./src

docker-build:
	docker build --no-cache=true --tag app:0.1 .

docker-run: docker-clean
	docker run --name app-0-1 --network=host app:0.1

docker-clean:
	-docker rm -f app-0-1