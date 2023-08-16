#!/usr/bin/env python3

import getopt
import re
import sys
import os
import tempfile
from pprint import pprint as pp
import subprocess


def helm_pull(path, version, local_helm_repo):
    subprocess.run(
        f"helm pull --untar --untardir {path} {local_helm_repo}/mccp  --version {version}",
        shell=True,
        check=True,
    )


def helm_template(path):
    return subprocess.run(
        f"helm template {path} --set policy-agent.enabled=true",
        shell=True,
        check=True,
        capture_output=True,
    ).stdout.decode("utf-8")


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
    return sorted({im for im in images if "kube-rbac-proxy" not in im})


def parse_image(image):
    image_repo_and_name, image_tag = image.split(":")
    image_name = image_repo_and_name.split("/")[-1]
    return image_name, image_tag


def to_acr_tag(image):
    image_name, image_tag = parse_image(image)
    return image_name + ':' + image_tag


def main(version, local_helm_repo, dry_run=False):
    with tempfile.TemporaryDirectory() as tmpdir:
        helm_pull(tmpdir, version, local_helm_repo)
        helm_template_output = helm_template(os.path.join(tmpdir, "mccp"))
        images = get_weaveworks_images(helm_template_output)
        for image in images:
            acr_repo = f"weaveworksmarketplacepublic.azurecr.io/"
            ecr_image = f"{acr_repo}{to_acr_tag(image)}"
            crane_copy(image, ecr_image, dry_run)


if __name__ == "__main__":
    help_text = f"""
Usage: {sys.argv[0]} [options]

e.g.
    {sys.argv[0]} --version 0.25.0 \\
        --local-helm-chart weave-gitops-enterprise-charts \\
        --dry-run

Options:
    -h, --help            show this help message and exit
    --dry-run             dry run
    --local-helm-chart    path to local helm chart
    --version             version of helm chart to publish
    """

    try:
        opts, args = getopt.getopt(
            sys.argv[1:],
            "h",
            ["help", "dry-run", "local-helm-chart=", "version="],
        )

    except getopt.GetoptError as err:
        print(str(err))
        print(help_text)
        sys.exit(2)

    dry_run = False
    local_helm_chart = None
    version = None

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

    if not version:
        print("version is required")
        print(help_text)
        sys.exit(2)

    if not local_helm_chart:
        print("local-helm-chart is required")
        print(help_text)
        sys.exit(2)

    main(version, local_helm_chart, dry_run)
