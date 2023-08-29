import { isAllowedLink } from '@weaveworks/weave-gitops';

export const openLinkHandler = (url: string) => {
  if (!isAllowedLink(url)) {
    return () => {};
  }
  return () => window.open(url, '_blank', 'noopener,noreferrer');
};
