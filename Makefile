BINARY_NAME=demo-service

build:
	go build -o ${BINARY_NAME} ./cmd/server/main.go

run: build
	./${BINARY_NAME} serve-user-http-api --config cfg/config.yaml

docker-up:
	docker-compose up -d --build

docker-down:
	docker-compose down

clean:
	go clean
	rm ${BINARY_NAME}

migrate-up:
	go run cmd/server/main.go migrate-db --config cfg/config.yaml

migrate-dry-run:
	go run cmd/server/main.go migrate-db --dry-run --config cfg/config.yaml

migrate-force:
	go run cmd/server/main.go migrate-db --force-migrate --config cfg/config.yaml

