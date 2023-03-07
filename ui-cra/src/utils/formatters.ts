import GitUrlParse from 'git-url-parse';
import { CostEstimate } from '../cluster-services/cluster_services.pb';
import { NotificationData } from '../contexts/Notifications';
import { URL } from '../types/global';
import { Routes } from './nav';

export const toPercent = (value: number, precision = 0) =>
  `${(100 * value).toFixed(precision)}%`;

export const trimTrailingForwardSlash = (url: URL) =>
  url.endsWith('/') ? url.slice(0, -1) : url;

export const intersperse = <T>(arr: T[], separator: (n: number) => T): T[] =>
  arr.reduce<T[]>((acc, currentElement, currentIndex) => {
    const isLast = currentIndex === arr.length - 1;
    return [
      ...acc,
      currentElement,
      ...(isLast ? [] : [separator(currentIndex)]),
    ];
  }, []);

export const getGitRepoHTTPSURL = (
  repoUrl?: string,
  repoBranch?: string,
): string => {
  if (repoUrl) {
    const parsedRepo = GitUrlParse(repoUrl);
    if (repoBranch) {
      return `https://${parsedRepo.resource}/${parsedRepo.full_name}/tree/${repoBranch}`;
    } else {
      return `https://${parsedRepo.resource}/${parsedRepo.full_name}`;
    }
  }
  return '';
};

export const getFormattedCostEstimate = (
  costEstimate: CostEstimate | undefined,
): string => {
  const costFormatter = new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
  });
  if (costEstimate) {
    const { currency, range } = costEstimate;
    const lowFormated = costFormatter.format(range?.low || 0);
    const highFormated = costFormatter.format(range?.high || 0);

    const estimate =
      (lowFormated === highFormated
        ? `${lowFormated}`
        : `${lowFormated} - ${highFormated}`) + ` ${currency}`;
    return estimate;
  } else return 'N/A';
};

export const formatError = (error: Error) =>
  [
    { message: { text: error.message }, severity: 'error' },
  ] as NotificationData[];

// Must be one of the valid URLs that we have already
// configured on the Gitlab backend for our Oauth app.
export function gitlabOAuthRedirectURI(): string {
  return `${window.location.origin}${Routes.GitlabOauthCallback}`;
}

export function bitbucketServerOAuthRedirectURI(): string {
  return `${window.location.origin}${Routes.BitBucketOauthCallback}`;
}
export function azureServerOAuthRedirectURI(): string {
  return `${window.location.origin}${Routes.AzureOauthCallback}`;
}
