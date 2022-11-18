import { Automation, Source } from '@weaveworks/weave-gitops/ui/lib/objects';
import { Pipeline } from '../../../api/pipelines/types.pb';
import { TerraformObject } from '../../../api/terraform/types.pb';
import { GitopsClusterEnriched } from '../../../types/custom';

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
) => {
  let annotation;
  if (resource.type === 'GitopsCluster') {
    annotation = (resource as GitopsClusterEnriched)?.annotations?.[
      'templates.weave.works/create-request'
    ];
  } else {
    annotation = (resource as Automation | Source)?.obj?.metadata
      ?.annotations?.['templates.weave.works/create-request'];
  }

  return maybeParseJSON(annotation || '');
};
