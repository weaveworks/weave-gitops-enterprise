import React, { useContext } from 'react';
import {
  request,
  requestWithCountHeader,
  requestWithEntitlementHeader,
} from '../../utils/request';

interface RequestContextType {
  fetch: typeof window.fetch;
}

const RequestContext = React.createContext<RequestContextType>(null as any);

// Injecting a fetch arg into our baseline request functions to allow for testing.
// This will allow us to simulate responses and error states and write tests against them.
const injectFetchArg = (fetch: Fetch, args: any) => {
  const fetchArg = args[3];
  if (fetchArg) {
    // The user has passed in their own fetch value here, so use that.
    return args;
  }

  return [args[0], args[1], args[2], fetch];
};

export type RequestHook = {
  request: typeof request;
  requestWithCountHeader: typeof requestWithCountHeader;
  requestWithEntitlementHeader: typeof requestWithEntitlementHeader;
};

export const useRequest = (): RequestHook => {
  const { fetch } = useContext(RequestContext);

  return {
    request: (...args: Parameters<typeof request>) => {
      console.log('called!');
      const next: Parameters<typeof request> = injectFetchArg(fetch, args);
      return request(...next);
    },
    requestWithCountHeader: (
      ...args: Parameters<typeof requestWithCountHeader>
    ) => {
      const next: Parameters<typeof requestWithCountHeader> = injectFetchArg(
        fetch,
        args,
      );
      return requestWithCountHeader(...next);
    },
    requestWithEntitlementHeader: (
      ...args: Parameters<typeof requestWithEntitlementHeader>
    ) => {
      const next: Parameters<typeof requestWithEntitlementHeader> =
        injectFetchArg(fetch, args);
      return requestWithEntitlementHeader(...next);
    },
  };
};

type Props = {
  fetch: typeof window.fetch;
  children: any;
};

export type Fetch = typeof window.fetch;

const RequestContextProvider = ({ fetch, children }: Props) => (
  <RequestContext.Provider value={{ fetch }}>
    {children}
  </RequestContext.Provider>
);

export default RequestContextProvider;
