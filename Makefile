BUILD_DIR=$(PWD)/bin

.PHONY: profile
profile:
	echo "Building profile service"
	mkdir -p ${BUILD_DIR} \
		&& cd backend_services/profile \
		&& GOOS=linux GOARCH=386 go build -o ${BUILD_DIR}/profile-service main.go

.PHONY: profile-test
profile-test:
	echo "Running unit tests for profile service"
	cd backend_services/profile && go test -short ./...

.PHONY: profile-test-integration
profile-test-integration:
	echo "Running integration tests for profile service"
	cd backend_services/profile \
		&& docker build . -t profile-service \
		&& mkdir -p test_data \
		&& grep -v '^--*' ../../schema/players.sql >test_data/schema.sql \
		&& go test --tags=integration ./...

.PHONY: matchmaking
matchmaking:
	echo "Building matchmaking service"
	mkdir -p ${BUILD_DIR} \
		&& cd backend_services/matchmaking \
		&& GOOS=linux GOARCH=386 go build -o ${BUILD_DIR}/matchmaking-service main.go

.PHONY: matchmaking-test
matchmaking-test:
	echo "Running unit tests for matchmaking service"
	cd backend_services/matchmaking && go test -short ./...

.PHONY: matchmaking-test-integration
matchmaking-test-integration:
	echo "Running integration tests for matchmaking service"
	cd backend_services/matchmaking \
		&& docker build . -t matchmaking-service \
		&& mkdir -p test_data \
		&& grep -v '^--*' ../../schema/players.sql >test_data/schema.sql \
		&& go test --tags=integration ./...

.PHONY: item
item:
	echo "Building item service"
	mkdir -p ${BUILD_DIR} \
		&& cd backend_services/item \
		&& GOOS=linux GOARCH=386 go build -o ${BUILD_DIR}/item-service main.go

.PHONY: item-test
item-test:
	echo "Running unit tests for item service"
	cd backend_services/item && go test ./...

.PHONY: item-test-integration
item-test-integration:
	echo "Running integration tests for item service"
	cd backend_services/item \
		&& docker build . -t item-service \
		&& mkdir -p test_data \
		&& grep -v '^--*' ../../schema/players.sql >test_data/schema.sql \
		&& echo ";" >> test_data/schema.sql \
		&& grep -v '^--*' ../../schema/trading.sql >> test_data/schema.sql \
		&& go test --tags=integration ./...

.PHONY: tradepost
tradepost:
	echo "Building tradepost service"
	mkdir -p ${BUILD_DIR} \
		&& cd backend_services/tradepost \
		&& GOOS=linux GOARCH=386 go build -o ${BUILD_DIR}/tradepost-service main.go

.PHONY: tradepost-test
tradepost-test:
	echo "Running unit tests for tradepost service"
	cd backend_services/tradepost && go test ./...

.PHONY: tradepost-test-integration
tradepost-test-integration:
	echo "Running integration tests for tradepost service"
	cd backend_services/tradepost \
		&& mkdir -p test_data \
		&& docker build . -t tradepost-service \
		&& grep -v '^--*' ../../schema/players.sql >test_data/schema.sql \
		&& echo ";" >> test_data/schema.sql \
		&& grep -v '^--*' ../../schema/trading.sql >> test_data/schema.sql \
		&& go test --tags=integration ./...

.PHONY: test-all-unit
test-all-unit: profile-test matchmaking-test item-test tradepost-test

.PHONY: test-all-integration
test-all-integration: profile-test-integration matchmaking-test-integration item-test-integration tradepost-test-integration

.PHONY: test-all
test-all: test-all-unit test-all-integration

.PHONY: build-all
build-all: profile matchmaking item tradepost

.PHONY: clean
clean:
	echo "Running cleanup"
	rm bin/*
	docker rmi -f profile-service matchmaking-service item-service tradepost-service
