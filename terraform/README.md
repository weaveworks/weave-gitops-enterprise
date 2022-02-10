### Services

#### Production
- https://gitlab.git.dev.weave.works running on cluster `gitlab-01` on GKE, backed by [repo](https://github.com/wkp-example-org/gitlab-01)



### How to run Terraform locally

1. Authenticate with GCP using `gcloud`.

```sh
gcloud auth application-default login
```

2. Switch to the working directory of your choice and run Terraform.
```sh
cd ./environments/dev
terraform init 
terraform plan
```

3. View output variables (optional).
```sh
terraform output
```
