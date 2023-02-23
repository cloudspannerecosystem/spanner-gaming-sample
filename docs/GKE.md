# GKE

This repository provides support for running the backend applications on GKE.

## Setup

### Terraform
The provided Terraform [gke.tf](../infrastructure/backend_gke.tf) will provision a [GKE Autopilot](https://cloud.google.com/kubernetes-engine/docs/concepts/autopilot-overview) cluster. This method of provisioning GKE will automatically manage the nodes for the cluster as the backend applications are added.

The terraform also enables the required cloud API services, such as `cloudbuild.googleapis.com` and `container.googleapis.com`.

Additionally, there is a Cloud Spanner service account created for the backend services.

### Kubectl
To interact with the GKE cluster, ensure kubectl is installed.

Once that is done, authenticate to GKE with the following commands:

```
export USE_GKE_GCLOUD_AUTH_PLUGIN=True
export GKE_CLUSTER=sample-game-gke # change this based on the terraform configuration
gcloud container clusters get-credentials $GKE_CLUSTER --region us-central1
kubectl get namespaces
```

If there are no issues with the kubectl commands, kubectl is properly authenticated.

### Create Cloud Build images
Each service and workload has a dockerfile that needs to be deployed to a container registry, such as [Google Container Registry](https://cloud.google.com/container-registry). The appropriate Cloud API services will have been enabled by the Terraform scripts found in the infrastructure folder.

To build the service containers, run the [`scripts/services_build.sh`](../scripts/services_build.sh) script. This will submit all service images to Cloud Build.

```
export PROJECT_ID=<YOUR_PROJECT_ID>
./scripts/services_build.sh
```

To build the workload containers, run the [`scripts/workloads_build.sh`](../scripts/workloads_build.sh) script. This will submit all service images to Cloud Build.

```
export PROJECT_ID=<YOUR_PROJECT_ID>
./scripts/workloads_build.sh
```

> **NOTE:** This will take some time time to complete all builds

### Spanner configuration and secrets

Once the images have been built, it is time to deploy the [kubernetes manifests](../kubernetes-manifests). Each backend application and workload provides a LoadBalance service and a deployment.

Create a config map for the Spanner instance, replace the variables with your Spanner instance and database information:

```
sed -e "s/PROJECT_ID/$PROJECT_ID/" \
    -e "s/INSTANCE_ID/$SPANNER_INSTANCE_ID/" \
    -e "s/DATABASE_ID/$SPANNER_DATABASE_ID/" \
    spanner_config.yaml.tmpl > ./kubernetes-manifests/spanner_config.yaml

kubectl apply -f spanner_config.yaml
```
> **NOTE:** [Workload Identity](https://cloud.google.com/kubernetes-engine/docs/concepts/workload-identity) manages credentials to ensure access to Cloud Spanner.

### Deploy the manifests
Once you have the kubernetes secret and config map established to connec to Cloud Spanner, the only thing left is to deploy the manifests.

To deploy the services, run the [`scripts/services_deploy.sh`](../scripts/services_deploy.sh).

```
export PROJECT_ID=<google cloud project id>
./scripts/services_deploy.sh
```

To deploy the workloads to GKE, run the [`scripts/workloads_deploy.sh`](../scripts/workloads_deploy.sh).

```
export PROJECT_ID=<google cloud project id>
./scripts/workloads_deploy.sh
```

For more information on running the workloads, follow [these instructions](./workloads.md).
