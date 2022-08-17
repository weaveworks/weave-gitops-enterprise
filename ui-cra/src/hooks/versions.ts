import { applicationsClient } from '@weaveworks/weave-gitops';
import { useContext, useEffect, useState } from 'react';
import { useQuery } from 'react-query';
import { GetConfigResponse } from '../cluster-services/cluster_services.pb';
import { EnterpriseClientContext } from '../contexts/EnterpriseClient';
import { useRequest } from '../contexts/Request';

export function useListVersion() {
  const { requestWithEntitlementHeader } = useRequest();
  return useQuery<{ data: any; entitlement: string | null }, Error>(
    'version',
    () => requestWithEntitlementHeader('GET', '/v1/enterprise/version'),
  );
}

export function useListConfig() {
  const { api } = useContext(EnterpriseClientContext);
  const [repoLink, setRepoLink] = useState<string>('');
  const queryResponse = useQuery<GetConfigResponse, Error>('config', () =>
    api.GetConfig({}),
  );
  const repositoryURL = queryResponse?.data?.repositoryURL || '';
  useEffect(() => {
    repositoryURL &&
      applicationsClient.ParseRepoURL({ url: repositoryURL }).then(res => {
        if (res.provider === 'GitHub') {
          setRepoLink(repositoryURL + `/pulls`);
        } else if (res.provider === 'GitLab') {
          setRepoLink(repositoryURL + `/-/merge_requests`);
        }
      });
  }, [repositoryURL]);

  return {
    ...queryResponse,
    repoLink,
  };
}
