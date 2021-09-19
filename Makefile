# @author: Roman Hzhyvinskyi
# @email: hzhyvinskyi@gmail.com

SERVICE_NAME := go-microservice-template
CONTAINER_PORT := 9000
HOST_PORT := 9000

docker-build-and-run:
	docker build --rm -t ${SERVICE_NAME} .
	docker run -it -p ${CONTAINER_PORT}:${HOST_PORT} ${SERVICE_NAME}

golangci-lint-run:
	golangci-lint run --no-config --disable-all --max-same-issues 500 --max-issues-per-linter 500 -v -E megacheck -E govet -E typecheck -E ineffassign

test:
	go test -race -v ./...

module-install: golangci-lint-run test docker-build-and-run
