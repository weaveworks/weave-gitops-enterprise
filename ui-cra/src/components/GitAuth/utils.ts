import { PageRoute } from '@weaveworks/weave-gitops/ui/lib/types';
import { GitProvider } from '../../api/gitauth/gitauth.pb';

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

const CALLBACK_STATE_KEY = 'oauth_callback_state';

export type CallbackSessionState = {
  page: PageRoute | string;
  state: any;
} | null;

export function storeCallbackState(value: CallbackSessionState) {
  if (!window.sessionStorage) {
    console.warn('no session storage found');
    return;
  }

  if (!value) {
    return null;
  }

  return sessionStorage.setItem(CALLBACK_STATE_KEY, JSON.stringify(value));
}

export function getCallbackState(): CallbackSessionState {
  const state = sessionStorage.getItem(CALLBACK_STATE_KEY);

  if (!state) {
    return null;
  }

  try {
    return JSON.parse(state);
  } catch (e) {
    return null;
  }
}

export function clearCallbackState() {
  return sessionStorage.removeItem(CALLBACK_STATE_KEY);
}

export enum GrpcErrorCodes {
  Unknown = 2,
  Unauthenticated = 16,
  NotFound = 5,
}

export const navigate = (url: string) => {
  if (process.env.NODE_ENV === 'test') {
    return;
  }
  window.location.href = url;
};
