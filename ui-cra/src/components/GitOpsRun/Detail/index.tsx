import { Page, RouterTab, SubRouterTabs } from '@weaveworks/weave-gitops';
import { useRouteMatch } from 'react-router-dom';
import styled from 'styled-components';
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
  const { path } = useRouteMatch();
  return (
    <Page path={[{ label: 'GitOps Run Detail' }]}>
      <PageTitle>{name}</PageTitle>
      <SubRouterTabs rootPath={`${path}/logs`}>
        <RouterTab name="Logs" path={`${path}/logs`}>
          <GitOpsRunLogs name={name || ''} namespace={namespace || ''} />
        </RouterTab>
      </SubRouterTabs>
    </Page>
  );
};

export default GitOpsRunDetail;
