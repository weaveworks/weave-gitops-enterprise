import { RequestMethod } from '../types/global';

export const processResponse = (res: Response) => {
  // 400s / 500s have res.ok = false
  if (!res.ok) {
    return res
      .clone()
      .json()
      .catch(() => res.text().then(message => ({ message })))
      .then(data => Promise.reject(data));
  }
  return res
    .clone()
    .json()
    .catch(() => res.text().then(message => ({ success: true, message })));
};

export const processResponseHeaders = (res: Response) => {
  const headersContent = res.headers.get('Content-Range')?.split('/');
  return headersContent?.[1];
};

export const request = (
  method: RequestMethod,
  query: RequestInfo,
  options: RequestInit = {},
) =>
  window.fetch(query, { ...options, method }).then(res => processResponse(res));

export const requestWithHeaders = (
  method: RequestMethod,
  query: RequestInfo,
  options: RequestInit = {},
) =>
  window.fetch(query, { ...options, method }).then(res =>
    processResponse(res).then(body => ({
      data: body,
      total: Number(processResponseHeaders(res)),
    })),
  );
