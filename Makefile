SHELL=/bin/bash -euo pipefail

# build images, test and run
.PHONY: prod-up
prod-up:
	@sudo ./scripts/run-prod.sh "up"

.PHONY: dev-up
dev-up:
	@sudo ./scripts/run-dev.sh "up"

# build images and test
.PHONY: prod-build
prod-build:
	@sudo ./scripts/run-prod.sh "build"

.PHONY: dev-build
dev-build:
	@sudo ./scripts/run-dev.sh "build"

# teardown running containers
.PHONY: prod-down
prod-down:
	@sudo ./scripts/run-prod.sh "down"

.PHONY: dev-down
dev-down:
	@sudo ./scripts/run-dev.sh "down"

# teardown running containers and delete all volumes
.PHONY: prod-down-volumes
prod-down-volumes:
	@sudo ./scripts/run-prod.sh "down-volumes"

.PHONY: dev-down-volumes
dev-down-volumes:
	@sudo ./scripts/run-dev.sh "down-volumes"

# test without building images
.PHONY: test
test: backend-test frontend-test

.PHONY: backend-test
backend-test:
	@cd ./backend; go test -v -race ./...

.PHONY: frontend-test
frontend-test:
	@cd ./frontend; npm run test

.PHONY: load
load:
	@cd ./backend/cmd/load; go run main.go 

.PHONY: clean
clean:
	@echo "removing compiled frontend..."
	@sudo rm -rf ./server/dist
	@echo "done"