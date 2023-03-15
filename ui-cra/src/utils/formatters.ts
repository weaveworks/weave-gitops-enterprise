import GitUrlParse from 'git-url-parse';
import { CostEstimate } from '../cluster-services/cluster_services.pb';
import { NotificationData } from '../contexts/Notifications';
import { URL } from '../types/global';
import { Routes } from './nav';
import { GitOpsSet } from '../api/gitopssets/types.pb';
import { TerraformObject } from '../api/terraform/types.pb';

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

export function azureDevOpsOAuthRedirectURI(): string {
  return `${window.location.origin}${Routes.AzureDevOpsOauthCallback}`;
}

export const getLabels = (
  obj: TerraformObject | GitOpsSet | undefined,
): [string, string][] => {
  const labels = obj?.labels;
  if (!labels) return [];
  return Object.keys(labels).flatMap(key => {
    return [[key, labels[key] as string]];
  });
};

const metadataPrefix = 'metadata.weave.works/';

export const getMetadata = (
  obj: TerraformObject | GitOpsSet | undefined,
): [string, string][] => {
  const annotations = obj?.annotations;
  if (!annotations) return [];
  return Object.keys(annotations).flatMap(key => {
    if (!key.startsWith(metadataPrefix)) {
      return [];
    } else {
      return [[key.slice(metadataPrefix.length), annotations[key] as string]];
    }
  });
};
