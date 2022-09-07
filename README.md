# Spanner Gaming Samples

This repository contains sample code for the following use-cases when using Cloud Spanner for the backend:

- Player creation and retrieval
- Basic game creation and matchmaking that tracks game statistics like games won and games played
- Item and currency acquisition for players in active games
- Ability to buy and sell items on a tradepost

## How to use this demo

### Setup infrastructure

#### Gcloud command line

To create the Spanner instance using gcloud, you must first [install and configure gcloud](https://cloud.google.com/sdk/docs/install-sdk).

When that's complete, ensure your gcloud project is set correctly.

```
gcloud config set project <PROJECT_ID>
```

> **NOTE:** You can find your PROJECT_ID in [Cloud Console](https://cloud.google.com/resource-manager/docs/creating-managing-projects#identifying_projects).

Now, create the Spanner instance and database:

```
gcloud spanner instances create game-instance --config=regional-us-central1 --description=gaming-instance --processing-units=500

gcloud spanner databases create --instance game-instance sample-game
```

> **NOTE:** The above command will create an instance using the us-central1 [regional configuration](https://cloud.google.com/spanner/docs/instance-configurations) with a compute capacity of 500 processing units. Be aware that creating an instance will start billing your account unless you are under Google Cloud's [free trial credits](https://cloud.google.com/free).

#### Terraform
A terraform file is provided that creates the appropriate resources for these samples.

Resources that are created:
- Spanner instance and database based on user variables in main.tfvars
- (TODO) GKE cluster to run the load generators

To set up the infrastructure, do the following:

- Copy `infrastructure/terraform.tfvars.sample` to `infrastructure/terraform.tfvars`
- Modify `infrastructure/terraform.tfvars` for PROJECT and instance configuration
- `terraform apply` from within infrastructure directory

```
cd infrastructure
cp terraform.tfvars.sample terraform.tfvars
vi terraform.tfvars # modify variables

terraform apply
```

### Setup schema
Schema is managed by [Wrench](https://github.com/cloudspannerecosystem/wrench).

After installing wrench, migrate the schema by running the `schema.bash` file (replace project/instance/database information with what was used in terraform file):

```
export SPANNER_PROJECT_ID=PROJECTID
export SPANNER_INSTANCE_ID=INSTANCEID
export SPANNER_DATABASE_ID=DATABASEID
./schema.bash
```

### Player profile sample

- Configure [`profile-service`](src/golang/profile-service) either by using environment variables or by copying the `profile-service/config.yml.template` file to `profile-service/config.yml`, and modify the Spanner connection details:

```
# environment variables
export SPANNER_PROJECT_ID=PROJECTID
export SPANNER_INSTANCE_ID=INSTANCEID
export SPANNER_DATABASE_ID=DATABASEID
```

```
# config.yml spanner connection details
spanner:
  project_id: YOUR_GCP_PROJECT_ID
  instance_id: YOUR_SPANNER_INSTANCE_ID
  database_id: YOUR_SPANNER_DATABASE_ID

```

- Run the profile service. By default, this will run the service on localhost:8080.

```
cd src/golang/profile-service
go run .
```

- Configure the [matchmaking-service](src/golang/matchmaking-service) either by using environment variables or by copying the `matchmaking-service/config.yml.template` file to `matchmaking-service/config.yml`, and modify the Spanner connection details:

```
# environment variables
export SPANNER_PROJECT_ID=PROJECTID
export SPANNER_INSTANCE_ID=INSTANCEID
export SPANNER_DATABASE_ID=DATABASEID
```

```
# config.yml spanner connection details
spanner:
  project_id: YOUR_GCP_PROJECT_ID
  instance_id: YOUR_SPANNER_INSTANCE_ID
  database_id: YOUR_SPANNER_DATABASE_ID

```

- Run the match-making service. By default, this will run the service on localhost:8081.

```
cd src/golang/matchmaking-service
go run .
```

- [Generate load](generators/README.md).

### Player trading sample

- Configure [`item-service`](src/golang/item-service) either by using environment variables or by copying the `item-service/config.yml.template` file to `item-service/config.yml`, and modify the Spanner connection details:

```
# environment variables
export SPANNER_PROJECT_ID=PROJECTID
export SPANNER_INSTANCE_ID=INSTANCEID
export SPANNER_DATABASE_ID=DATABASEID
```

```
# config.yml spanner connection details
spanner:
  project_id: YOUR_GCP_PROJECT_ID
  instance_id: YOUR_SPANNER_INSTANCE_ID
  database_id: YOUR_SPANNER_DATABASE_ID

```

- Run the item service. By default, this will run the service on localhost:8082.

```
cd src/golang/item-service
go run .
```

- Configure the [tradepost-service](src/golang/tradepost-service) either by using environment variables or by copying the `tradepost-service/config.yml.template` file to `tradepost-service/config.yml`, and modify the Spanner connection details:

```
# environment variables
export SPANNER_PROJECT_ID=PROJECTID
export SPANNER_INSTANCE_ID=INSTANCEID
export SPANNER_DATABASE_ID=DATABASEID
```

```
# config.yml spanner connection details
spanner:
  project_id: YOUR_GCP_PROJECT_ID
  instance_id: YOUR_SPANNER_INSTANCE_ID
  database_id: YOUR_SPANNER_DATABASE_ID

```

- Run the tradepost service. By default, this will run the service on localhost:8083.

```
cd src/golang/tradepost-service
go run .
```

- [Generate load](generators/README.md).


### Generator dependencies

The generators are run by Locust.io, which is a Python framework for generating load.

There are several dependencies required to get the generators to work:

- Python 3.7+
- Locust

Assuming python3.X is installed, install dependencies via [pip](https://pypi.org/project/pip/):

```
# if pip3 is symlinked to pip
pip install -r requirements.txt

# if pip3 is not symlinked to pip
pip3 install -r requirements.txt
```

> **NOTE:** To avoid modifying existing pip libraries on your machine, consider a solution like [virtualenv](https://pypi.org/project/virtualenv/).

## How to build the services

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

## How to clean up

### GCloud command line

If the Spanner instance was created using the gcloud command line, it can be delete using gcloud:

```
gcloud spanner instances delete game-instance
```

### Terraform

If the Spanner instance was created using terraform, then from the `infrastructure` directory you can destroy the infrastructure.

```
cd infrastructure
terraform destroy
```

### Clean up build and tests
The Makefile provides a `make clean` command that removes the binaries and docker containers that were created as part of building and testing the services.

```
make clean
```
