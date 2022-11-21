import { Automation, Source } from '@weaveworks/weave-gitops/ui/lib/objects';
import { Pipeline } from '../../../api/pipelines/types.pb';
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
    | TerraformObject
    | Pipeline,
  yaml?: string,
) => {
  const getAnnotation = (resourceType: string) => {
    switch (resourceType) {
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
      case 'Pipeline':
        return yamlConverter.load(yaml).metadata.annotations[
          'templates.weave.works/create-request'
        ];
      default:
        return;
    }
  };

  return maybeParseJSON(getAnnotation(resource.type || ''));
};
