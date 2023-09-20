import { Page } from '../../Layout/App';
import { NotificationsWrapper } from '../../Layout/NotificationsWrapper';
import GitOpsRunLogs from './GitOpsRunLogs';
import { RouterTab, SubRouterTabs } from '@weaveworks/weave-gitops';
import { useRouteMatch } from 'react-router-dom';
import styled from 'styled-components';

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
      <NotificationsWrapper>
        <PageTitle>{name}</PageTitle>
        <SubRouterTabs rootPath={`${path}/logs`}>
          <RouterTab name="Logs" path={`${path}/logs`}>
            <GitOpsRunLogs name={name || ''} namespace={namespace || ''} />
          </RouterTab>
        </SubRouterTabs>
      </NotificationsWrapper>
    </Page>
  );
};

export default GitOpsRunDetail;
