import React, { FC, useCallback, useContext, useEffect, useState } from 'react';
import { processEntitlementHeaders } from '../../utils/request';
import { Versions, VersionData } from './index';
import useNotifications from './../Notifications';
import { useHistory } from 'react-router-dom';
import { EnterpriseClientContext } from '../EnterpriseClient';

const VersionsProvider: FC = ({ children }) => {
  const [entitlement, setEntitlement] = useState<string | null>(null);
  const [versions, setVersions] = useState<VersionData>({
    ui: process.env.REACT_APP_VERSION || 'no version specified',
  });
  const [repositoryURL, setRepositoryURL] = useState<string | null>(null);
  const { setNotifications } = useNotifications();
  const { api } = useContext(EnterpriseClientContext);

  const history = useHistory();

  const getVersions = useCallback(() => {
    // requestWithEntitlementHeader('GET', '/v1/enterprise/version', {
    //   cache: 'no-store',
    // })
    api
      .GetEnterpriseVersion({})
      .then((res: any) => {
        return {
          data: res,
          entitlement: processEntitlementHeaders(res),
        };
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
  }, [api, setNotifications]);

  const getConfig = useCallback(() => {
    api
      .GetConfig({})
      .then((res: any) => setRepositoryURL(res.repositoryURL))
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
