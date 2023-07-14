#!/usr/bin/env python3

import re
import sys
import subprocess

from publish_images_azure import helm_template


def get_all_images(yaml_stream):
    images = []
    for line in yaml_stream.splitlines():
        groups = re.search(r"image: (.*)", line)
        if groups:
            images.append(groups.group(1))
    return sorted({im for im in images})


def docker_pull_image(image, dry_run):
    cmd = f"docker pull {image}"

    if dry_run:
        print(cmd)
        return

    subprocess.run(cmd, shell=True, check=True)

def get_tagged_image(image):
    image_ref, sha = image.split('@')
    return f"{image_ref}:to-kind"


def docker_tag(image, dry_run):
    tagged_image = get_tagged_image(image)
    cmd = f"docker tag {image} {tagged_image}"

    if dry_run:
        print(cmd)
        return

    subprocess.run(cmd, shell=True, check=True)


def kind_load_image(image, dry_run):
    tagged_image = get_tagged_image(image)
    cmd = f"kind load docker-image {tagged_image}"

    if dry_run:
        print(cmd)
        return

    subprocess.run(cmd, shell=True, check=True)


def main(helm_chart_path, dry_run):
    helm_template_output = helm_template(helm_chart_path)
    images = get_all_images(helm_template_output)
    for image in images:
        docker_pull_image(image, dry_run)
        docker_tag(image, dry_run)
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
