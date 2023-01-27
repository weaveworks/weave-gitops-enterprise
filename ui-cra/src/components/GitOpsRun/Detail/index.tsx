import { RouterTab, SubRouterTabs } from '@weaveworks/weave-gitops';
import { useRouteMatch } from 'react-router-dom';
import styled from 'styled-components';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';
import { useGetLogs } from '../../../hooks/gitopsrun';
import GitOpsRunLogs from './GitOpsRunLogs';
type Props = {
  name?: string;
  namespace?: string;
  clusterName?: string;
};

const PageTitle = styled.h4`
  font-size: ${({ theme }) => theme.fontSizes.large};
  font-weight: 600;
  margin: 0;
  margin-bottom: ${({ theme }) => theme.spacing.small};
`;

const GitOpsRunDetail = ({ name, namespace, clusterName }: Props) => {
  const { isLoading, data, error } = useGetLogs({
    namespace,
    sessionId: 'run-master-fdd0a0b7',
    clusterName: 'default/run-master-fdd0a0b7',
  });

  // Previous args (which displayed the same error)

  // const { isLoading, data, error } = useGetLogs({
  //   namespace,
  //   sessionId: name,
  //   clusterName,
  // });

  console.log('logs:');
  console.log(data);

  const fakes = [
    {
      source: 'bucket-server',
      message: 'Cleanup Bucket Source and Kustomization successfully',
      severity: 'info',
      timestamp: '2022-08-14 12:20:00 UTC',
    },
    {
      source: 'bucket-server',
      message: 'Cleanup Bucket Source and Kustomization successfully',
      severity: 'error',
      timestamp: '2022-08-14 12:20:00 UTC',
    },
  ];
  const { path } = useRouteMatch();
  return (
    <PageTemplate
      documentTitle="GitOps Run Detail"
      path={[{ label: 'GitOps Run Detail' }]}
    >
      <ContentWrapper
      // loading={isLoading}
      // errors={[{ message: error?.message }]}
      >
        <PageTitle>{name}</PageTitle>
        <SubRouterTabs rootPath={`${path}/logs`}>
          <RouterTab name="Logs" path={`${path}/logs`}>
            <GitOpsRunLogs logs={fakes || []} />
          </RouterTab>
        </SubRouterTabs>
      </ContentWrapper>
    </PageTemplate>
  );
};

export default GitOpsRunDetail;
