import {
  ListConfigResponse,
  useListConfig,
  useListVersion,
} from '../../hooks/versions';
import React from 'react';

export const ListConfigContext = React.createContext<ListConfigResponse | null>(
  null,
);

export const ListConfigProvider = ({ children }: { children: any }) => {
  const listConfig = useListConfig();
  return (
    <ListConfigContext.Provider value={listConfig}>
      {children}
    </ListConfigContext.Provider>
  );
};

export const useListConfigContext = () => React.useContext(ListConfigContext);

// Use react contexct to share the Version in footer component to save some renders
export const VersionContext = React.createContext<
  | {
      data: any;
      entitlement: string | null;
    }
  | undefined
>(undefined);

export const VersionProvider = ({ children }: { children: any }) => {
  const { data } = useListVersion();

  return (
    <VersionContext.Provider value={data}>{children}</VersionContext.Provider>
  );
};

export const useVersionContext = () => React.useContext(VersionContext);
