import { FC, useCallback, useContext, useEffect, useState } from 'react';
import { useHistory } from 'react-router-dom';
import { EnterpriseClientContext } from '../EnterpriseClient';
import { useRequest } from '../Request';
import useNotifications from './../Notifications';
import { VersionData, Versions } from './index';

const VersionsProvider: FC = ({ children }) => {
  const [entitlement, setEntitlement] = useState<string | null>(null);
  const [versions, setVersions] = useState<VersionData>({
    ui: process.env.REACT_APP_VERSION || 'no version specified',
  });
  const [repositoryURL, setRepositoryURL] = useState<string | null>(null);
  const { setNotifications } = useNotifications();
  const { api } = useContext(EnterpriseClientContext);
  const { requestWithEntitlementHeader } = useRequest();
  console.log(requestWithEntitlementHeader);

  const history = useHistory();

  const getVersions = useCallback(() => {
    requestWithEntitlementHeader('GET', '/v1/enterprise/version', {
      cache: 'no-store',
    })
      .then(res => {
        setVersions(s => ({ ...s, capiServer: res.data.version }));
        setEntitlement(res.entitlement);
      })
      .catch(err =>
        setNotifications([
          { message: { text: err.message }, variant: 'danger' },
        ]),
      );
  }, [setNotifications, requestWithEntitlementHeader]);

  const getConfig = useCallback(() => {
    api
      .GetConfig({})
      .then(res => setRepositoryURL(res.repositoryURL as string))
      .catch(err =>
        setNotifications([
          { message: { text: err.message }, variant: 'danger' },
        ]),
      );
  }, [api, setNotifications]);

  useEffect(() => {
    getVersions();
    getConfig();

    return history.listen(() => {
      getVersions();
      getConfig();
    });
  }, [getVersions, getConfig, history]);

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
