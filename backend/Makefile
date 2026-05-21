include .env
export $(shell sed 's/=.*//' .env)

run: generate-docs
	go run ./cmd/api/main.go

build:
	go build -o much-to-do ./cmd/api/main.go
	mv much-to-do ./bin/much-to-do

clean:
	rm -f ./bin/much-to-do

tidy:
	go mod tidy

dc-up:
	docker-compose -f docker-compose.yaml up -d --build

dc-restart:
	docker-compose -f docker-compose.yaml restart

# confirm-rs:
# # 	docker exec -it mongodb mongosh -u ${MONGO_INITDB_ROOT_USERNAME} -p ${MONGO_INITDB_ROOT_PASSWORD} --authenticationDatabase admin --eval "rs.status()"
# 	docker exec mongodb mongosh \
#   		-u "${MONGO_INITDB_ROOT_USERNAME:-root}" \
#   		-p "${MONGO_INITDB_ROOT_PASSWORD:-example}" \
#   		--authenticationDatabase admin \
#   		--eval 'if (rs.conf() === null) { rs.initiate({_id: "rs0", members: [{_id: 0, host: "mongodb:27017"}]}) } else { print("Replica set already configured.") }'
# 	docker exec -it mongodb mongosh --u ${MONGO_INITDB_ROOT_USERNAME} -p ${MONGO_INITDB_ROOT_PASSWORD} --authenticationDatabase admin --eval "rs.status()"

dc-down:
	docker-compose -f docker-compose.yaml down

generate-docs:
	swag init -g ./cmd/api/main.go -o ./docs

unit-test:
	go test -v ./...

integration-test:
	go test -tags=integration -v ./...

.PHONY: run build clean tidy dc-up dc-down generate-docs unit-test integration-test
