# Local deployment

The following instructions highlight how to run the backend_services locally.

> **NOTE:** Skip these sections if you deployed the services using GKE.

## Local player profile deployment

- Configure [`profile-service`](./backend_services/profile-service) either by using environment variables or by copying the `profile-service/config.yml.template` file to `profile-service/config.yml`, and modify the Spanner connection details:

```
# environment variables. change the YOUR_* values to your information
export SPANNER_PROJECT_ID=YOUR_PROJECT_ID
export SPANNER_INSTANCE_ID=YOUR_INSTANCE_ID
export SPANNER_DATABASE_ID=YOUR_DATABASE_ID
```

```
# config.yml spanner connection details. change the YOUR_* values to your information
spanner:
  project_id: YOUR_PROJECT_ID
  instance_id: YOUR_INSTANCE_ID
  database_id: YOUR_DATABASE_ID

```

- Run the profile service. By default, this will run the service on localhost:8080.

```
cd ./backend_services/profile-service
go run .
```

- Configure the [matchmaking-service](./backend_services/matchmaking-service) either by using environment variables or by copying the `matchmaking-service/config.yml.template` file to `matchmaking-service/config.yml`, and modify the Spanner connection details:

```
# environment variables
export SPANNER_PROJECT_ID=YOUR_PROJECT_ID
export SPANNER_INSTANCE_ID=YOUR_INSTANCE_ID
export SPANNER_DATABASE_ID=YOUR_DATABASE_ID
```

```
# config.yml spanner connection details
spanner:
  project_id: YOUR_PROJECT_ID
  instance_id: YOUR_INSTANCE_ID
  database_id: YOUR_DATABASE_ID

```

- Run the match-making service. By default, this will run the service on localhost:8081.

```
cd ./backend_services/matchmaking-service
go run .
```

## Local player trading deployment

- Configure [`item-service`](./backend_services/item-service) either by using environment variables or by copying the `item-service/config.yml.template` file to `item-service/config.yml`, and modify the Spanner connection details:

```
# environment variables
export SPANNER_PROJECT_ID=YOUR_PROJECT_ID
export SPANNER_INSTANCE_ID=YOUR_INSTANCE_ID
export SPANNER_DATABASE_ID=YOUR_DATABASE_ID
```

```
# config.yml spanner connection details
spanner:
  project_id: YOUR_PROJECT_ID
  instance_id: YOUR_INSTANCE_ID
  database_id: YOUR_DATABASE_ID

```

- Run the item service. By default, this will run the service on localhost:8082.

```
cd ./backend_services/item-service
go run .
```

- Configure the [tradepost-service](./backend_services/tradepost-service) either by using environment variables or by copying the `tradepost-service/config.yml.template` file to `tradepost-service/config.yml`, and modify the Spanner connection details:

```
# environment variables
export SPANNER_PROJECT_ID=YOUR_PROJECT_ID
export SPANNER_INSTANCE_ID=YOUR_INSTANCE_ID
export SPANNER_DATABASE_ID=YOUR_DATABASE_ID
```

```
# config.yml spanner connection details
spanner:
  project_id: YOUR_PROJECT_ID
  instance_id: YOUR_INSTANCE_ID
  database_id: YOUR_DATABASE_ID

```

- Run the tradepost service. By default, this will run the service on localhost:8083.

```
cd ./backend_services/tradepost-service
go run .
```

## Workloads

Once the services are deployed you can use the Locust generators to [run workloads](./docs/workloads.md).

## How to build the services locally

A Makefile is provided to build the services. Example commands:

```
# Build everything
make build-all

# Build individual services
make profile
make matchmaking
make item
make tradepost
```

> **NOTE:** The build command currently assumes GOOS=linux and GOARCH=386. Building on other platforms currently is not supported.

## How to run the service tests
A Makefile is provided to test the services. Both unit tests and integration tests are provided.

Example commands:

```
make profile-test
make profile-test-integration

make test-all-unit
make test-all-integration

make test-all
```

> **NOTE:** The tests rely on [testcontainers-go](https://github.com/testcontainers/testcontainers-go), so [Docker](https://www.docker.com/) must be installed.
