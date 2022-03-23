### Services

#### Production
- https://gitlab.git.dev.weave.works running on cluster `gitlab-01` on GKE, backed by [repo](https://github.com/wkp-example-org/gitlab-01)

#### Development

- https://dex-01.wge.dev.weave.works running on cluster `dex-01` on GKE, backed by [repo](https://github.com/wkp-example-org/dex-01). OIDC issuer configured for Github and on-prem Gitlab.
- https://demo-01.wge.dev.weave.works running on cluster `demo-01` on GKE, backed by [repo](https://gitlab.git.dev.weave.works/wge/demo-01). WGE demo cluster, manually updated.
- https://demo-02.wge.dev.weave.works running on cluster `demo-02` on GKE, backed by [repo](https://github.com/wkp-example-org/demo-02). WGE demo cluster, manually updated.

### How to run Terraform locally

1. Authenticate with GCP using `gcloud`:
```sh
gcloud auth application-default login
```

2. Authenticate with AWS using `aws-google-auth`:
```sh
docker run -it -e GOOGLE_USERNAME -e GOOGLE_IDP_ID -e GOOGLE_SP_ID -e AWS_DEFAULT_REGION -e AWS_PROFILE -e AWS_ROLE_ARN -v ~/.aws:/root/.aws cevoaustralia/aws-google-auth --resolve-aliases

```

3. Switch to the working directory of your choice and run Terraform:
```sh
cd ./environments/dev
terraform init 
terraform plan
```

4. View output variables (optional):
```sh
terraform output
```
