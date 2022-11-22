import { Automation, Source } from '@weaveworks/weave-gitops/ui/lib/objects';
import { Pipeline } from '../../../api/pipelines/types.pb';
import { GetTerraformObjectResponse } from '../../../api/terraform/terraform.pb';
import { GitopsClusterEnriched } from '../../../types/custom';

const yamlConverter = require('js-yaml');

export const maybeParseJSON = (data: string) => {
  try {
    return JSON.parse(data);
  } catch (e) {
    // FIXME: show a warning to a user or something
    return undefined;
  }
};

export const getCreateRequestAnnotation = (
  resource:
    | GitopsClusterEnriched
    | Automation
    | Source
    | GetTerraformObjectResponse
    | Pipeline,
) => {
  const type =
    (resource as GitopsClusterEnriched | Automation | Source | Pipeline).type ||
    (resource as GetTerraformObjectResponse)?.object?.type ||
    '';

  console.log(
    type,
    (resource as Automation | Source)?.obj?.metadata?.annotations?.[
      'templates.weave.works/create-request'
    ],
  );

  const getAnnotation = (
    resource:
      | GitopsClusterEnriched
      | Automation
      | Source
      | GetTerraformObjectResponse
      | Pipeline,
    type: string,
  ) => {
    switch (type) {
      case 'GitopsCluster':
        return (resource as GitopsClusterEnriched)?.annotations?.[
          'templates.weave.works/create-request'
        ];
      case 'GitRepository':
      case 'Bucket':
      case 'HelmRepository':
      case 'HelmChart':
      case 'Kustomization':
      case 'HelmRelease':
      case 'OCIRepository':
        return (resource as Automation | Source)?.obj?.metadata?.annotations?.[
          'templates.weave.works/create-request'
        ];
      case 'Terraform':
        return yamlConverter.load((resource as GetTerraformObjectResponse).yaml)
          .metadata.annotations['templates.weave.works/create-request'];
      // add case for Pipeline
      default:
        return '';
    }
  };

  return maybeParseJSON(getAnnotation(resource, type));
};
