import { GitopsClusterEnriched } from '../../../types/custom';

export const maybeParseJSON = (data: string) => {
  try {
    return JSON.parse(data);
  } catch (e) {
    // FIXME: show a warning to a user or something
    return undefined;
  }
};

export const getCreateRequestAnnotation = (resource: any) => {
  return (
    resource?.annotations &&
    maybeParseJSON(
      resource?.annotations['templates.weave.works/create-request'],
    )
  );
};
