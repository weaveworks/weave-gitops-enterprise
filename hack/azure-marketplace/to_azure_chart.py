import json
import os
import re
import subprocess
import sys
import yaml
import deepmerge


def find_yaml_files(directory):
    for root, dirs, files in os.walk(directory):
        for file in files:
            if file.endswith(".yaml"):
                yield os.path.join(root, file)


unique_parts = set()


def extract_deployment_name_from_path(path):
    path_parts = os.path.normpath(path).split(os.sep)
    if 'charts' in path_parts:
        return path_parts[path_parts.index('charts') + 1]
    else:
        return path_parts[path_parts.index('templates') + 1]


def replace_image_string(doc, part_of_value):
    image_pattern = r'(image:\s*)(.*(\n\s*\|.*)?)'

    def image_replacement(match):
        image_field = match.group(0)
        imageName = part_of_value
        if 'kubeRbacProxy' in image_field or 'kube-rbac-proxy' in image_field:
            imageName = "kubeRbacProxy"
        elif "uiServer" in image_field:
            imageName = "uiServer"

        unique_parts.add(imageName)

        return fr'{match.group(1)}{{{{ .Values.global.azure.images.{imageName}.registry }}}}/{{{{ .Values.global.azure.images.{imageName}.image }}}}@{{{{ .Values.global.azure.images.{imageName}.digest }}}}'

    return re.sub(image_pattern, image_replacement, doc, flags=re.MULTILINE)


def modify_yaml_file(file):
    with open(file, 'r') as f:
        content = f.read()

    # split the file into documents
    documents = re.split(r'\n---\n', content)

    for i, doc in enumerate(documents):
        # check if the document is a Deployment
        if re.search(r'kind:\s*Deployment', doc):
            part_of_value = extract_deployment_name_from_path(file)

            # convert kebab-case to camelCase, e.g. azure-vote-back to azureVoteBack
            part_of_value = ''.join([word.capitalize()
                                    for word in part_of_value.split('-')])
            part_of_value = part_of_value[0].lower() + part_of_value[1:]

            # update image field in templates
            doc = replace_image_string(doc, part_of_value)

            # add azure-extensions-usage-release-identifier label to each pod template
            label_pattern = r'(template:\s*\n\s*metadata:\s*\n\s*labels:\s*\n)'
            label_replacement = r'\1        azure-extensions-usage-release-identifier: {{ .Release.Name }}\n'
            doc = re.sub(label_pattern, label_replacement, doc)

            # update the document in the list
            documents[i] = doc

    # join the documents back together
    content = '\n---\n'.join(documents)

    with open(file, 'w') as f:
        f.write(content)


# define your list of docker repos
repos = [
    "cluster-bootstrap-controller",
    "cluster-controller",
    "gitopssets-controller",
    "kube-rbac-proxy",
    "pipeline-controller",
    "policy-agent",
    "templates-controller",
    "weave-gitops-enterprise-clusters-service",
    "weave-gitops-enterprise-ui-server",
]

# Azure settings
REGISTRY_NAME = 'weaveworksmarketplacepublic.azurecr.io'


def get_latest_sha256(repo):
    try:
        # fetch the manifest of the most recent tag to get the digest
        output = subprocess.check_output(
            ['az', 'acr', 'manifest', 'list-metadata', '--output', 'json', f"{REGISTRY_NAME}/{repo}"])
        manifests = json.loads(output)

        manifests_with_tags = [
            manifest for manifest in manifests if manifest.get('tags')
        ]

        if not manifests_with_tags:
            print(f"Error fetching information for {repo}: no tags found")
            print(f"Manifests: {manifests}")
            return None, None

        manifests_with_tags = sorted(
            manifests_with_tags, key=lambda manifest: manifest['createdTime'])

        most_recent_manifest = manifests_with_tags[-1]

        return most_recent_manifest["tags"][0], most_recent_manifest['digest']
    except Exception as e:
        print(f"Error fetching information for {repo}: {e}")
        return None, None

    return None, None


def build_latest_shas():
    data = {}

    for repo in repos:
        tag, digest = get_latest_sha256(repo)
        print(f"{repo}: {tag} {digest}")
        data[repo] = digest

    return data


def build_values_yaml():

    shas = build_latest_shas()

    data_values_yaml = f"""
global:
  azure:
    images:
      clusterBootstrapController:
        digest: {shas['cluster-bootstrap-controller']}
        image: cluster-bootstrap-controller
        registry: weaveworksmarketplacepublic.azurecr.io
      clusterController:
        digest: {shas['cluster-controller']}
        image: cluster-controller
        registry: weaveworksmarketplacepublic.azurecr.io
      clustersService:
        digest: {shas['weave-gitops-enterprise-clusters-service']}
        image: weave-gitops-enterprise-clusters-service
        registry: weaveworksmarketplacepublic.azurecr.io
      gitopssetsController:
        digest: {shas['gitopssets-controller']}
        image: gitopssets-controller
        registry: weaveworksmarketplacepublic.azurecr.io
      kubeRbacProxy:
        digest: {shas['kube-rbac-proxy']}
        image: kube-rbac-proxy
        registry: weaveworksmarketplacepublic.azurecr.io
      pipelineController:
        digest: {shas['pipeline-controller']}
        image: pipeline-controller
        registry: weaveworksmarketplacepublic.azurecr.io
      policyAgent:
        digest: {shas['policy-agent']}
        image: policy-agent
        registry: weaveworksmarketplacepublic.azurecr.io
      templatesController:
        digest: {shas['templates-controller']}
        image: templates-controller
        registry: weaveworksmarketplacepublic.azurecr.io
      uiServer:
        digest: {shas['weave-gitops-enterprise-ui-server']}
        image: weave-gitops-enterprise-ui-server
        registry: weaveworksmarketplacepublic.azurecr.io
"""

    data_values = yaml.safe_load(data_values_yaml)
    return data_values


def main(helm_chart_directory):

    data_values = build_values_yaml()

    for file in find_yaml_files(helm_chart_directory):
        modify_yaml_file(file)

        values = {
            "global": {
                "azure": {
                    "images": {}
                }
            }
        }

    for part_of in unique_parts:
        if part_of not in data_values["global"]["azure"]["images"]:
            raise Exception(f"Part of {part_of} not found in data_values.yaml")
        values["global"]["azure"]["images"][part_of] = data_values["global"]["azure"]["images"][part_of]

    # read the values.yaml file
    with open(os.path.join(helm_chart_directory, 'values.yaml'), 'r') as f:
        content = f.read()

    # merge the values.yaml file with the values dictionary
    original_values = yaml.safe_load(content)
    # deep merge the dictionaries, we'll use the library "merge" from the "deepmerge" package
    original_values = deepmerge.always_merger.merge(original_values, values)

    # write the values.yaml file
    with open(os.path.join(helm_chart_directory, 'values.yaml'), 'w') as f:
        yaml.dump(original_values, f)


def test_replace_image_string_single():
    doc = """
    image: {{ .Values.controller.manager.image.repository }}:{{ .Values.controller.manager.image.tag | default .Chart.AppVersion }}
    """
    result = replace_image_string(doc, "backend")
    expected = """
    image: {{ .Values.global.azure.images.backend.registry }}/{{ .Values.global.azure.images.backend.image }}@{{ .Values.global.azure.images.backend.digest }}
    """
    assert expected.strip() == result.strip()


def test_replace_image_string():
    doc = """
    image: {{ .Values.controller.manager.image.repository }}:{{ .Values.controller.manager.image.tag
        | default .Chart.AppVersion }}
    """
    result = replace_image_string(doc, "backend")
    expected = """
    image: {{ .Values.global.azure.images.backend.registry }}/{{ .Values.global.azure.images.backend.image }}@{{ .Values.global.azure.images.backend.digest }}
    """
    assert expected.strip() == result.strip()


def test_replace_image_string_with_surrounding_fields():
    doc = """
    name: my-app
    image: {{ .Values.controller.manager.image.repository }}:{{ .Values.controller.manager.image.tag
        | default .Chart.AppVersion }}
    replicas: 3
    """
    result = replace_image_string(doc, "backend")
    expected = """
    name: my-app
    image: {{ .Values.global.azure.images.backend.registry }}/{{ .Values.global.azure.images.backend.image }}@{{ .Values.global.azure.images.backend.digest }}
    replicas: 3
    """
    assert expected.strip() == result.strip()


def test_replace_image_string_with_templated_surrounding_fields():
    doc = """
    name: {{ .Values.app.name }}
    image: {{ .Values.controller.manager.image.repository }}:{{ .Values.controller.manager.image.tag
        | default .Chart.AppVersion }}
    replicas: {{ .Values.app.replicas }}
    """
    result = replace_image_string(doc, "backend")
    expected = """
    name: {{ .Values.app.name }}
    image: {{ .Values.global.azure.images.backend.registry }}/{{ .Values.global.azure.images.backend.image }}@{{ .Values.global.azure.images.backend.digest }}
    replicas: {{ .Values.app.replicas }}
    """
    assert expected.strip() == result.strip()


if __name__ == "__main__":
    # read path from command line
    path = sys.argv[1]
    print(main(path))
