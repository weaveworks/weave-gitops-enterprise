import React, { FC, useCallback, useEffect, useState } from 'react';
import { ClustersService } from '../../capi-server/capi_server.pb';
import { Versions, VersionData } from './index';

const VersionsProvider: FC = ({ children }) => {
  const [versions, setVersions] = useState<VersionData>({
    ui: process.env.REACT_APP_VERSION || 'no version specified',
  });

  const getVersions = useCallback(() => {
    ClustersService.GetEnterpriseVersion({}).then(res => {
      console.log('weave-gitops-enterprise capiServer:', res.version);
      setVersions(s => ({ ...s, capiServer: res.version }));
    });
  }, []);

  useEffect(() => getVersions(), [getVersions]);

  return (
    <Versions.Provider
      value={{
        versions,
      }}
    >
      {children}
    </Versions.Provider>
  );
};

export default VersionsProvider;
