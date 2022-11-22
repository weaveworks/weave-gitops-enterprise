import { Automation, Source } from '@weaveworks/weave-gitops/ui/lib/objects';
import { Pipeline } from '../../../api/pipelines/types.pb';
import { GetTerraformObjectResponse } from '../../../api/terraform/terraform.pb';
import { TerraformObject } from '../../../api/terraform/types.pb';
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
      case 'Source':
      case 'Automation':
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

  console.log(type);
  console.log(
    yamlConverter.load((resource as GetTerraformObjectResponse).yaml).metadata
      .annotations['templates.weave.works/create-request'],
  );

  return maybeParseJSON(getAnnotation(resource, type));
};
