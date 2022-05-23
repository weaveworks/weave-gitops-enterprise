import { RequestMethod } from '../types/global';

export const processResponse = (res: any) => {
  // 400s / 500s have res.ok = false
  if (!res.ok) {
    return res
      .clone()
      .json()
      .catch(() => res.text().then((message: any) => ({ message })))
      .then((data: any) => Promise.reject(data));
  }
  return res
    .clone()
    .json()
    .catch(() =>
      res.text().then((message: any) => ({ success: true, message })),
    );
};

export const processCountHeader = (res: any) => {
  const headersContent = res.headers.get('Content-Range')?.split('/');
  return headersContent?.[1];
};

export const processEntitlementHeaders = (res: Response) => {
  return res.headers?.get('Entitlement-Expired-Message');
};

export const request = (
  method: RequestMethod,
  query: RequestInfo,
  options: RequestInit = {},
) =>
  window.fetch(query, { ...options, method }).then(res => processResponse(res));

export const requestWithCountHeader = (
  method: RequestMethod,
  query: RequestInfo,
  options: RequestInit = {},
) =>
  window.fetch(query, { ...options, method }).then(res =>
    processResponse(res).then((body: any) => ({
      data: body,
      total: Number(processCountHeader(res)),
    })),
  );

export const requestWithEntitlementHeader = (
  method: RequestMethod,
  query: RequestInfo,
  options: RequestInit = {},
) =>
  window.fetch(query, { ...options, method }).then(res =>
    processResponse(res).then((body: any) => ({
      data: body,
      entitlement: processEntitlementHeaders(res),
    })),
  );

export enum GrpcErrorCodes {
  Unauthenticated = 16,
}

export const isUnauthenticated = (code: number): boolean => {
  return code === GrpcErrorCodes.Unauthenticated;
};

export const removeToken = (provider: string) =>
  localStorage.removeItem(`gitProviderToken_${provider}`);
