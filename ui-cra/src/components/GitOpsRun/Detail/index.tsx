import { RouterTab, SubRouterTabs } from '@weaveworks/weave-gitops';
import { useEffect, useState } from 'react';
import { useRouteMatch } from 'react-router-dom';
import styled from 'styled-components';
import { useGetLogs } from '../../../hooks/gitopsrun';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';
import GitOpsRunLogs from './GitOpsRunLogs';
type Props = {
  name?: string;
  namespace?: string;
};

const PageTitle = styled.h4`
  font-size: ${({ theme }) => theme.fontSizes.large};
  font-weight: 600;
  margin: 0;
  margin-bottom: ${({ theme }) => theme.spacing.small};
`;

const GitOpsRunDetail = ({ name, namespace }: Props) => {
  const [token, setToken] = useState('');
  const { isLoading, data, error } = useGetLogs({
    sessionNamespace: namespace,
    sessionId: name,
    token,
  });

  console.log(data);

  useEffect(() => {
    if (isLoading) return;
    setToken(data?.nextToken || '');
  }, [data]);

  const { path } = useRouteMatch();
  return (
    <PageTemplate
      documentTitle="GitOps Run Detail"
      path={[{ label: 'GitOps Run Detail' }]}
    >
      <ContentWrapper
        loading={isLoading}
        errors={[{ message: error?.message }]}
      >
        <PageTitle>{name}</PageTitle>
        <SubRouterTabs rootPath={`${path}/logs`}>
          <RouterTab name="Logs" path={`${path}/logs`}>
            <GitOpsRunLogs logs={data?.logs || []} />
          </RouterTab>
        </SubRouterTabs>
      </ContentWrapper>
    </PageTemplate>
  );
};

export default GitOpsRunDetail;
