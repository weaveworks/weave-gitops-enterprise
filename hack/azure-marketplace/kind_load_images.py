#!/usr/bin/env python3

import sys
import subprocess

from publish_images_azure import get_weaveworks_images, helm_template


def docker_pull_image(image, dry_run):
    cmd = f"docker pull {image}"

    if dry_run:
        print(cmd)
        return

    subprocess.run(cmd, shell=True, check=True)


def kind_load_image(image, dry_run):
    cmd = f"kind load docker-image {image} --name kind-wge-dev"

    if dry_run:
        print(cmd)
        return

    subprocess.run(cmd, shell=True, check=True)


def main(helm_chart_path, dry_run):
    helm_template_output = helm_template(helm_chart_path)
    images = get_weaveworks_images(helm_template_output)
    for image in images:
        docker_pull_image(image, dry_run)
        kind_load_image(image, dry_run)


if __name__ == "__main__":

    help_text = """
    Usage: python kind_load_images.py <helm_chart_path> [--dry-run]
    """
    if len(sys.argv) < 2:
        print(help_text)
        sys.exit(1)

    helm_chart_path = sys.argv[1]
    dry_run = False
    if len(sys.argv) > 2:
        dry_run = sys.argv[2] == "--dry-run"

    main(helm_chart_path, dry_run)
