import { GetVersionResponse } from '@weaveworks/progressive-delivery';
import { useContext, useEffect, useState } from 'react';
import { useQuery } from 'react-query';
import { GetConfigResponse } from '../cluster-services/cluster_services.pb';
import { useAPI } from '../contexts/API';
import { GitAuth } from '../contexts/GitAuth';
import { useRequest } from '../contexts/Request';

export function useListVersion() {
  const { requestWithEntitlementHeader } = useRequest();
  return useQuery<
    { data: GetVersionResponse; entitlement: string | null },
    Error
  >('version', () =>
    requestWithEntitlementHeader('GET', '/v1/enterprise/version'),
  );
}
export interface ListConfigResponse extends GetConfigResponse {
  uiConfig: any;
  [key: string]: any;
}

export function useListConfig() {
  const { enterprise } = useAPI();
  const [provider, setProvider] = useState<string>('');
  const queryResponse = useQuery<GetConfigResponse, Error>('config', () =>
    enterprise.GetConfig({}),
  );
  const { gitAuthClient } = useContext(GitAuth);

  const repositoryURL = queryResponse?.data?.repositoryUrl || '';
  const uiConfig = JSON.parse(queryResponse?.data?.uiConfig || '{}');
  useEffect(() => {
    repositoryURL &&
      gitAuthClient.ParseRepoURL({ url: repositoryURL }).then(res => {
        setProvider(res.provider || '');
      });
  }, [repositoryURL, gitAuthClient]);

  return {
    ...queryResponse,
    uiConfig,
    provider,
  };
}
