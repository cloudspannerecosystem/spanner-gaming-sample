# GKE

This repository provides support for running the backend applications on GKE.

## Setup

### Terraform
The provided Terraform [gke.tf](../infrastructure/gke.tf) will provision a [GKE Autopilot](https://cloud.google.com/kubernetes-engine/docs/concepts/autopilot-overview) cluster. This method of provisioning GKE will automatically manage the nodes for the cluster as the backend applications are added.

These terraform [services.tf](../infrastructure/services.tf) file enables the required cloud API services, such as `cloudbuild.googleapis.com` and `container.googleapis.com`.

Additionally, there is a Cloud Spanner service account created for the backend services.

### Kubectl
To interact with the GKE cluster, ensure kubectl is installed.

Once that is done, authenticate to GKE with the following commands:

```
export USE_GKE_GCLOUD_AUTH_PLUGIN=True
export GKE_CLUSTER=cymbal-games-gke # change this based on the terraform configuration
gcloud container clusters get-credentials $GKE_CLUSTER --region us-central1
kubectl get namespaces
```

If there are no issues with the kubectl commands, kubectl is properly authenticated.

### Create Cloud Build images
Each service has a dockerfile that needs to be deployed to a container registry, such as [Google Container Registry](https://cloud.google.com/container-registry). The appropriate Cloud API services will have been enabled by the Terraform scripts found in the infrastructure folder.

To deploy the services, run the [`scripts/build.sh`](../scripts/build.sh) script. This will submit all service images to Cloud Build.

```
export PROJECT_ID=<google cloud project id>
./scripts/build.sh
```

> **NOTE:** This will take some time time to complete all builds

### Spanner configuration and secrets

Once the images have been built, it is time to deploy the [kubernetes manifests](../kubernetes-manifests). Each backend application provides a LoadBalance service and a deployment.

Create the secret for the service account that will connect to the Spanner instance:

```
export SERVICE_ACCOUNT=cymbal-games-backend
gcloud iam service-accounts keys create backend_sa_key.json \
    --iam-account=${SERVICE_ACCOUNT}@${PROJECT_ID}.iam.gserviceaccount.com

kubectl create secret generic cymbal-games-backend-sa-key \
 --from-file=backend_sa_key.json=./backend_sa_key.json
```

Create a config map for the Spanner instance

```
sed -e "s/PROJECT_ID/$PROJECT_ID/" \
    -e "s/INSTANCE_ID/$SPANNER_INSTANCE_ID/" \
    -e "s/DATABASE_ID/$SPANNER_DATABASE_ID/" \
    spanner_config.yaml.tmpl > spanner_config.yaml

kubectl apply -f spanner_config.yaml
```

### Deploy the manifests
Once you have the kubernetes secret and config map established to connec to Cloud Spanner, the only thing left is to deploy the manifests. A [`scripts/deploy.sh`](../scripts/deploy.sh) file has been created to assist with this process.

```
export PROJECT_ID=<google cloud project id>
./scripts/deploy.sh
```
