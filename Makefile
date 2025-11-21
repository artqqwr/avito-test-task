SERVER_MAIN_PATH = ./cmd/server/main.go
BIN_PATH = ./bin
OAPI_CODEGEN_CONFIG_PATH = ./api/oapi-codegen.yaml
OAPI_SPEC_PATH = ./api/openapi.yaml
K6_LOAD_TEST_PATH = ./load_test.js

generate:
	@echo "Generating Go code from openapi spec"
	@go tool oapi-codegen -config $(OAPI_CODEGEN_CONFIG_PATH) $(OAPI_SPEC_PATH)

run: generate
	@echo "Running $(SERVER_MAIN_PATH)"
	@go run $(SERVER_MAIN_PATH)

build: generate
	@echo "Building $(SERVER_MAIN_PATH)"
	@go build -o $(BIN_PATH)/server $(SERVER_MAIN_PATH)
	@echo "Builded into $(BIN_PATH)"
	@echo "Running $(BIN_PATH)/server"

compose-up:
	docker compose up -d

compose-down:
	docker compose down

# idk how to make this correctly
compose-clean:
	@echo "Cleaning all docker-compose images and containers"
	docker compose down --rmi all

clean: compose-clean
	@echo "Cleaning $(BIN_PATH)"
	@rm -rf $(BIN_PATH)

load-test:
	@k6 run $(K6_LOAD_TEST_PATH)