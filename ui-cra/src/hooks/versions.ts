import { useContext, useEffect, useState } from 'react';
import { useQuery } from 'react-query';
import { GetConfigResponse } from '../cluster-services/cluster_services.pb';
import { EnterpriseClientContext } from '../contexts/EnterpriseClient';
import { useRequest } from '../contexts/Request';
import GitUrlParse from 'git-url-parse';

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
    if (repositoryURL) {
      const { resource, full_name, protocol } = GitUrlParse(repositoryURL);
      const provider = resource?.split('.')[0];
      if (provider === 'github') {
        setRepoLink(`https://github.com/${full_name}/pulls`);
      } else if (provider === 'gitlab') {
        setRepoLink(`${protocol}://${resource}/${full_name}/-/merge_requests`);
      }
    }
  }, [repositoryURL]);

  return {
    ...queryResponse,
    repoLink,
  };
}
