import { URL } from '../types/global';
import GitUrlParse from 'git-url-parse';

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
    return `https://${parsedRepo.resource}/${parsedRepo.full_name}/tree/${repoBranch}`;
  }
  return '';
};
