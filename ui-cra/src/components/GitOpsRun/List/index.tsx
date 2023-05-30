import { Flex, Page, useListObjects } from '@weaveworks/weave-gitops';
import GitOpsRunTable from './GitOpsRunTable';
import NoRunsMessage from './NoRunsMessage';

const GitOpsRun = () => {
  const { data: sessions, isLoading } = useListObjects('', 'StatefulSet', '', {
    app: 'vcluster',
    'app.kubernetes.io/part-of': 'gitops-run',
  });
  return (
    <Page
      loading={isLoading}
      error={sessions?.errors}
      path={[{ label: 'GitOps Run' }]}
    >
      {sessions?.objects?.length ? (
        <GitOpsRunTable sessions={sessions.objects} />
      ) : (
        <Flex center wide>
          <NoRunsMessage />
        </Flex>
      )}
    </Page>
  );
};

export default GitOpsRun;
