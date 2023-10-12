import { Flex, useListObjects } from '@weaveworks/weave-gitops';
import { Page } from '../../Layout/App';
import { NotificationsWrapper } from '../../Layout/NotificationsWrapper';
import GitOpsRunTable from './GitOpsRunTable';
import NoRunsMessage from './NoRunsMessage';

const GitOpsRun = () => {
  const { data: sessions, isLoading } = useListObjects('', 'StatefulSet', '', {
    app: 'vcluster',
    'app.kubernetes.io/part-of': 'gitops-run',
  });
  return (
    <Page loading={isLoading} path={[{ label: 'GitOps Run' }]}>
      <NotificationsWrapper errors={sessions?.errors}>
        {sessions?.objects?.length ? (
          <GitOpsRunTable sessions={sessions.objects} />
        ) : (
          <Flex center>
            <NoRunsMessage />
          </Flex>
        )}
      </NotificationsWrapper>
    </Page>
  );
};

export default GitOpsRun;
