# Spanner Gaming Samples

This repository contains sample code for the following use-cases when using Cloud Spanner for the backend:

- Player creation and retrieval
- Basic game creation and matchmaking that tracks game statistics like games won and games played
- Item and currency acquisition for players in active games
- Ability to buy and sell items on a tradepost

![gaming_backend_services.png](images/gaming_backend_services.png)

## REST endpoints

These are the REST endpoints exposed by backend services

### profile service
![gaming_backend_profile_rest.png](images/gaming_backend_profile_rest.png)

### matchmaking service
![gaming_backend_matchmaking_rest.png](images/gaming_backend_matchmaking_rest.png)

### item service
![gaming_backend_item_rest.png](images/gaming_backend_item_rest.png)

### tradepost service
![gaming_backend_tradepost_rest.png](images/gaming_backend_tradepost_rest.png)

## Spanner schema

The Cloud Spanner schema that supports the backend services looks like this.

### Players and games
![gaming_schema_players_and_games.png](images/gaming_schema_players_and_games.png)

### Items and player ledger
![gaming_schema_items.png](images/gaming_schema_items.png)

> **NOTE:** Players table is repeated to show relation

### Trades and player items
![gaming_schema_trades.png](images/gaming_schema_trades.png)

> **NOTE:** Players and player_items tables are repeated to show relations

## How to use this demo

### Setup infrastructure

You can either set up the Spanner infrastructure using the gcloud command line or Terraform. Instructions for both are below.

> **NOTE:** The Terraform scripts also create a GKE Autopilot cluster.

Before you set up the infrastructure, it is important to enable the appropriate APIs using the gcloud command line.

You must [install and configure gcloud](https://cloud.google.com/sdk/docs/install-sdk).

When that's complete, ensure your gcloud project is set correctly.

```
gcloud config set project <PROJECT_ID>
```

> **NOTE:** You can find your PROJECT_ID in [Cloud Console](https://cloud.google.com/resource-manager/docs/creating-managing-projects#identifying_projects).

Then you can set up the Spanner infrastructure using either the gcloud command line or Terraform. Instructions for both are below.

> **NOTE:** The Terraform scripts also create a GKE Autopilot cluster.

#### Gcloud command line

To create the Spanner instance and database using gcloud, issue the following commands:

```
gcloud spanner instances create sample-instance --config=regional-us-central1 --description=gaming-instance --processing-units=500

gcloud spanner databases create --instance sample-instance sample-game
```

> **NOTE:** The above command will create an instance using the us-central1 [regional configuration](https://cloud.google.com/spanner/docs/instance-configurations) with a compute capacity of 500 processing units. Be aware that creating an instance will start billing your account unless you are under Google Cloud's [free trial credits](https://cloud.google.com/free).

#### Terraform
A terraform file is provided that creates the appropriate resources for these samples.

Resources that are created:
- Spanner instance and database based on user variables in main.tfvars
- [GKE cluster](./docs/GKE.md) to run the services

To set up the infrastructure, do the following:

- Copy `infrastructure/terraform.tfvars.sample` to `infrastructure/terraform.tfvars`
- Modify `infrastructure/terraform.tfvars` for PROJECT and instance configuration
- `terraform apply` from within infrastructure directory

```
cd infrastructure
terraform init
cp terraform.tfvars.sample terraform.tfvars
vi terraform.tfvars # modify variables

# Authenticate to gcloud services so Terraform can make changes
gcloud auth application-default login

terraform apply
```

### Schema management
Schema is managed by [Wrench](https://github.com/cloudspannerecosystem/wrench).

After installing wrench, migrate the schema by running the `./scripts/schema.sh` file (replace project/instance/database information with what was used in terraform file):

```
export SPANNER_PROJECT_ID=YOUR_PROJECT_ID
export SPANNER_INSTANCE_ID=YOUR_INSTANCE_ID
export SPANNER_DATABASE_ID=YOUR_DATABASE_ID
./scripts/schema.sh
```

> **NOTE:** The schema must be in place for the services to work. Do not skip this step!

### Deploy services
You can deploy the services to the GKE cluster that was configured by Terraform, or you can deploy them locally.

To deploy to GKE, follow the [instructions here](./docs/GKE.md).

Otherwise, follow the local deployment instructions for player profile and tradepost.

Once the services are deployed you can use the generators to [run workloads](./docs/workloads.md).

Then follow the README to clean up based on whether you deployed with gcloud or Terraform.

#### Local player profile deployment

> **NOTE:** Skip this section if you deployed the services using [GKE](./docs/GKE.md)

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

#### Local player trading deployment

> **NOTE:** Skip this section if you deployed the services using [GKE](./docs/GKE.md)

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

## Cleaning up

### GCloud command line

If the Spanner instance was created using the gcloud command line, it can be delete using gcloud:

```
gcloud spanner instances delete sample-instance
```

### Terraform

If the infrastructure was created using terraform, then from the `infrastructure` directory you can destroy the infrastructure.

```
cd infrastructure
terraform destroy
```

### Clean up builds and tests
The Makefile provides a `make clean` command that removes the binaries and docker containers that were created as part of building and testing the services.

```
make clean
```
