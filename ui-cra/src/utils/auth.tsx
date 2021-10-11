//@ts-nocheck

const tokenKey = (providerName: string) => `gitProviderToken_${providerName}`;

export function storeProviderToken(providerName: string, token: string) {
  if (!window.localStorage) {
    console.warn('no local storage found');
    return;
  }

  localStorage.setItem(tokenKey(providerName), token);
}

export function getProviderToken(providerName: string): string {
  if (!window.localStorage) {
    console.warn('no local storage found');
    return;
  }

  return localStorage.getItem(tokenKey(providerName));
}
