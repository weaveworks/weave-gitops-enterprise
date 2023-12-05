import { GetVersionResponse } from '@weaveworks/progressive-delivery';
import { withBasePath } from '@weaveworks/weave-gitops';
import { useEffect, useState } from 'react';
import { useQuery } from 'react-query';
import { GetConfigResponse } from '../cluster-services/cluster_services.pb';
import { useEnterpriseClient } from '../contexts/API';
import { useRequest } from '../contexts/Request';

export function useListVersion() {
  const { requestWithEntitlementHeader } = useRequest();
  return useQuery<
    { data: GetVersionResponse; entitlement: string | null },
    Error
  >('version', () =>
    requestWithEntitlementHeader('GET', withBasePath('/v1/enterprise/version')),
  );
}
export interface ListConfigResponse extends GetConfigResponse {
  uiConfig: any;
  [key: string]: any;
}

export function useListConfig() {
  const { clustersService } = useEnterpriseClient();
  const [provider, setProvider] = useState<string>('');
  const queryResponse = useQuery<GetConfigResponse, Error>('config', () =>
    clustersService.GetConfig({}),
  );
  const { gitAuth } = useEnterpriseClient();

  const repositoryURL = queryResponse?.data?.repositoryUrl || '';
  const uiConfig = JSON.parse(queryResponse?.data?.uiConfig || '{}');
  useEffect(() => {
    repositoryURL &&
      gitAuth.ParseRepoURL({ url: repositoryURL }).then(res => {
        setProvider(res.provider || '');
      });
  }, [repositoryURL, gitAuth]);

  return {
    ...queryResponse,
    uiConfig,
    provider,
  };
}
