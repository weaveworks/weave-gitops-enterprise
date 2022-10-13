import { GitopsClusterEnriched } from '../../../types/custom';

export const maybeParseJSON = (data: string) => {
  try {
    return JSON.parse(data);
  } catch (e) {
    // FIXME: show a warning to a user or something
    return undefined;
  }
};

export const getCreateRequestAnnotation = (annotation: any) =>
  maybeParseJSON(annotation);
