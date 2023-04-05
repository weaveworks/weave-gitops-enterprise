#!/usr/bin/env python3

import getopt
import re
import sys
import os
import tempfile
from pprint import pprint as pp
import subprocess

AWS_ECR_REPO = "709825985650.dkr.ecr.us-east-1.amazonaws.com/weaveworks"
AWS_IMAGE_NAMES = [
    "weave-gitops-enterprise-development",
    "weave-gitops-enterprise-production",
]


def helm_pull(path, version, local_helm_repo):
    subprocess.run(
        f"helm pull --untar --untardir {path} {local_helm_repo}/mccp  --version {version}",
        shell=True,
        check=True,
    )


def helm_template(path):
    return subprocess.run(
        f"helm template {path}",
        shell=True,
        check=True,
        capture_output=True,
    ).stdout.decode("utf-8")


def helm_package(path):
    subprocess.run(
        f"helm package {path}/mccp",
        shell=True,
        check=True,
    )


def crane_copy(src, dst, dry_run):
    cmd = f"crane cp {src} {dst}"

    if dry_run:
        print(cmd)
        return

    subprocess.run(
        cmd,
        shell=True,
        check=True,
    )


def get_weaveworks_images(yaml_stream):
    images = []
    for line in yaml_stream.splitlines():
        groups = re.search(r"image: (.*)", line)
        if groups:
            images.append(groups.group(1))
    return sorted({im for im in images if "weaveworks" in im})


def parse_image(image):
    image_repo_and_name, image_tag = image.split(":")
    image_name = image_repo_and_name.split("/")[-1]
    return image_name, image_tag


def to_ecr_tag(image):
    image_name, image_tag = parse_image(image)
    return f"{image_name}-{image_tag}"


def update_chart_values(path, image_values):
    with tempfile.NamedTemporaryFile(mode="w") as f:
        f.write(image_values)
        f.flush()
        return subprocess.run(
            f"yq --inplace '. *= load(\"{f.name}\")' {path}/mccp/values.yaml",
            shell=True,
            check=True,
        )


def update_chart_name(path, aws_product):
    return subprocess.run(
        f"yq --inplace '.name = \"{aws_product}\"' {path}/mccp/Chart.yaml",
        shell=True,
        check=True,
    )


def helm_push_oci_chart(aws_image_name, version, dry_run):
    cmd = f"helm push ./{aws_image_name}-{version}.tgz oci://{AWS_ECR_REPO}"

    if dry_run:
        print(cmd)
        return

    return subprocess.run(
        cmd,
        shell=True,
        check=True,
    )


def get_values(images, ecr_repo):
    images_to_update = {}
    for image in images:
        image_name, _ = parse_image(image)
        images_to_update[image_name] = to_ecr_tag(image)

    template = f"""
images:
  clustersService: {ecr_repo}:{images_to_update["weave-gitops-enterprise-clusters-service"]}
  uiServer: {ecr_repo}:{images_to_update["weave-gitops-enterprise-ui-server"]}
  clusterBootstrapController: {ecr_repo}:{images_to_update["cluster-bootstrap-controller"]}

cluster-controller:
  controllerManager:
    manager:
      image:
        repository: {ecr_repo}
        tag: {images_to_update["cluster-controller"]}

pipeline-controller:
  controller:
    manager:
      image:
        repository: {ecr_repo}
        tag: {images_to_update["pipeline-controller"]}

templates-controller:
  controllerManager:
    manager:
      image:
        repository: {ecr_repo}
        tag: {images_to_update["templates-controller"]}

gitopssets-controller:
  controllerManager:
    manager:
      image:
        repository: {ecr_repo}
        tag: {images_to_update["gitopssets-controller"]}
""".strip()
    return template


def main(version, aws_image_name, local_helm_repo, dry_run=False):
    with tempfile.TemporaryDirectory() as tmpdir:
        helm_pull(tmpdir, version, local_helm_repo)
        helm_template_output = helm_template(os.path.join(tmpdir, "mccp"))
        images = get_weaveworks_images(helm_template_output)
        ecr_repo = f"{AWS_ECR_REPO}/{aws_image_name}"
        for image in images:
            ecr_image = f"{ecr_repo}:{to_ecr_tag(image)}"
            crane_copy(image, ecr_image, dry_run)
        image_values = get_values(images, ecr_repo)
        update_chart_values(tmpdir, image_values)
        update_chart_name(tmpdir, aws_image_name)
        helm_package(tmpdir)
        helm_push_oci_chart(aws_image_name, version, dry_run)


if __name__ == "__main__":
    help_text = f"""
Usage: {sys.argv[0]} [options]

e.g.
    {sys.argv[0]} --version 0.20.0 \\
        --aws-image-name weave-gitops-enterprise-development \\
        --local-helm-chart weave-gitops-enterprise-charts \\
        --dry-run

Options:
    -h, --help            show this help message and exit
    --dry-run             dry run
    --local-helm-chart    path to local helm chart
    --version             version of helm chart to publish
    --aws-image-name      name of AWS image to publish to
    """

    try:
        opts, args = getopt.getopt(
            sys.argv[1:],
            "h",
            ["help", "dry-run", "local-helm-chart=", "version=", "aws-image-name="],
        )

    except getopt.GetoptError as err:
        print(str(err))
        print(help_text)
        sys.exit(2)

    dry_run = False
    local_helm_chart = None
    version = None
    aws_image_name = None

    for opt, arg in opts:
        if opt in ("-h", "--help"):
            print(help_text)
            sys.exit()
        elif opt == "--dry-run":
            dry_run = True
        elif opt == "--local-helm-chart":
            local_helm_chart = arg
        elif opt == "--version":
            version = arg
        elif opt == "--aws-image-name":
            aws_image_name = arg

    if not version:
        print("version is required")
        print(help_text)
        sys.exit(2)

    if aws_image_name not in AWS_IMAGE_NAMES:
        print(f"aws-image-name must be one of {AWS_IMAGE_NAMES}")
        print(help_text)
        sys.exit(2)

    if not local_helm_chart:
        print("local-helm-chart is required")
        print(help_text)
        sys.exit(2)

    main(version, aws_image_name, local_helm_chart, dry_run)
