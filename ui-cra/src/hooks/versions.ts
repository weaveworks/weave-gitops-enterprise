import { useContext, useEffect, useState } from 'react';
import { useQuery } from 'react-query';
import { GetConfigResponse } from '../cluster-services/cluster_services.pb';
import { EnterpriseClientContext } from '../contexts/EnterpriseClient';
import { useRequest } from '../contexts/Request';
import GitUrlParse from 'git-url-parse';
import { GitAuth } from '../contexts/GitAuth';

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
  const { gitAuthClient } = useContext(GitAuth);

  const repositoryURL = queryResponse?.data?.repositoryURL || '';
  const uiConfig = JSON.parse(queryResponse?.data?.uiConfig || '{}');
  useEffect(() => {
    repositoryURL &&
      gitAuthClient.ParseRepoURL({ url: repositoryURL }).then(res => {
        const { resource, full_name, protocol } = GitUrlParse(repositoryURL);
        if (res.provider === 'GitHub') {
          setRepoLink(`${protocol}://${resource}/${full_name}/pulls`);
        } else if (res.provider === 'GitLab') {
          setRepoLink(
            `${protocol}://${resource}/${full_name}/-/merge_requests`,
          );
        }
      });
  }, [repositoryURL, gitAuthClient]);

  return {
    ...queryResponse,
    repoLink,
    uiConfig,
  };
}
