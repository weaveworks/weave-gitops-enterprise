export const maybeParseJSON = (data: string) => {
  try {
    return JSON.parse(data);
  } catch (e) {
    // FIXME: show a warning to a user or something
    return undefined;
  }
};

export const getCreateRequestAnnotation = (resource: any, type?: string) => {
  let annotation: string | undefined = '';
  if (type === 'Cluster') {
    annotation =
      resource?.annotations &&
      resource.annotations['templates.weave.works/create-request'];
  } else {
    annotation =
      resource?.obj?.metadata?.annotations &&
      resource?.obj?.metadata?.annotations[
        'templates.weave.works/create-request'
      ];
  }
  return annotation && maybeParseJSON(annotation);
};
