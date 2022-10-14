import { Automation, Source } from '@weaveworks/weave-gitops/ui/lib/objects';
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
  resource: GitopsClusterEnriched | Automation | Source,
) => {
  const annotation =
    resource.type === 'Cluster'
      ? (resource as GitopsClusterEnriched)?.annotations?.[
          'templates.weave.works/create-request'
        ]
      : (resource as Automation | Source)?.obj?.annotations?.[
          'templates.weave.works/create-request'
        ];
  return maybeParseJSON(annotation);
};
