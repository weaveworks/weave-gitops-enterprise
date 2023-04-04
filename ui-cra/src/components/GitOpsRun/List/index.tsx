import { Flex, useListObjects } from '@weaveworks/weave-gitops';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';

import GitOpsRunTable from './GitOpsRunTable';
import NoRunsMessage from './NoRunsMessage';

const GitOpsRun = () => {
  const { data: sessions, isLoading } = useListObjects('', 'StatefulSet', '', {
    app: 'vcluster',
    'app.kubernetes.io/part-of': 'gitops-run',
  });
  return (
    <PageTemplate documentTitle="GitOps Run" path={[{ label: 'GitOps Run' }]}>
      <ContentWrapper loading={isLoading} errors={sessions?.errors}>
        {sessions?.objects?.length ? (
          <GitOpsRunTable sessions={sessions.objects} />
        ) : (
          <Flex center>
            <NoRunsMessage />
          </Flex>
        )}
      </ContentWrapper>
    </PageTemplate>
  );
};

export default GitOpsRun;
