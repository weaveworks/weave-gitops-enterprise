import React, { FC, useCallback, useEffect, useState } from 'react';
import { request, requestWithEntitlementHeader } from '../../utils/request';
import { Versions, VersionData } from './index';
import useNotifications from './../Notifications';

const VersionsProvider: FC = ({ children }) => {
  const [entitlement, setEntitlement] = useState<string | null>(null);
  const [versions, setVersions] = useState<VersionData>({
    ui: process.env.REACT_APP_VERSION || 'no version specified',
  });
  const [repositoryURL, setRepositoryURL] = useState<string | null>(null);
  const { setNotifications } = useNotifications();

  const getVersions = useCallback(() => {
    requestWithEntitlementHeader('GET', '/v1/enterprise/version', {
      cache: 'no-store',
    })
      .then(res => {
        setVersions(s => ({ ...s, capiServer: res.data.version }));
        setEntitlement(res.entitlement);
      })
      .catch(err =>
        setNotifications([{ message: err.message, variant: 'danger' }]),
      );
  }, [setNotifications]);

  const getConfig = useCallback(() => {
    request('GET', '/v1/config', {
      cache: 'no-store',
    })
      .then(res => {
        setRepositoryURL(res.setRepositoryURL);
      })
      .catch(err => {
        setRepositoryURL('https://gitlab.com/user/blog');
        setNotifications([{ message: err.message, variant: 'danger' }]);
      });
  }, [setNotifications]);

  useEffect(() => {
    getVersions();
    getConfig();
  }, [getVersions, getConfig]);

  return (
    <Versions.Provider
      value={{
        versions,
        entitlement,
        repositoryURL,
      }}
    >
      {children}
    </Versions.Provider>
  );
};

export default VersionsProvider;
