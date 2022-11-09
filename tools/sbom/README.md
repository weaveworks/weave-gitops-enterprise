# Sbom for wge

## Goal 

- able to generate sbom for all components within wge
- different formats of container images

## What

some requirements 

1. easy to generate for a given wge environment 
2. easy to consume it  
3. easy to run on demand or scheduled. Ideally it runs on demand for each release of wg ee / wg oss

## Alternatives
There are some potential candidates

- https://github.com/ckotzbauer/sbom-operator
- https://aquasecurity.github.io/trivy/v0.34/docs/kubernetes/operator/
- https://github.com/kubernetes-sigs/bom


### Sbom-operator

- https://github.com/ckotzbauer/sbom-operator

#### how to generate for a given wge environment


#### how to consume generated sboms

#### how to run on demand on scheduled

- It can run as scheduled job via cron or 
- It could run as controller 




