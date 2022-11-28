export enum GitProvider {
  Unknown = 'Unknown',
  GitHub = 'GitHub',
  GitLab = 'GitLab',
}
export interface GetGithubDeviceCodeResponse {
  userCode?: string | undefined;
  deviceCode?: string | undefined;
  validationURI?: string | undefined;
  interval?: number | undefined;
}
export interface GetGithubAuthStatusResponse {
  accessToken?: string | undefined;
  error?: string | undefined;
}

const tokenKey = (providerName: GitProvider) =>
  `gitProviderToken_${providerName}`;

export function storeProviderToken(providerName: GitProvider, token: string) {
  if (!window.localStorage) {
    console.warn('no local storage found');
    return;
  }
  localStorage.setItem(tokenKey(providerName), token);
}

export function getProviderToken(providerName: GitProvider): string | null {
  if (!window.localStorage) {
    console.warn('no local storage found');
    return null;
  }
  return localStorage.getItem(tokenKey(providerName));
}
