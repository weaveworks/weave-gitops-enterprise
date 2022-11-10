import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';

import { GitOpsRunTable } from './GitOpsRunTable';
import NoRunsMessage from './NoRunsMessage';

const GitOpsRun = () => {
  const sessions: any[] = [];
  return (
    <PageTemplate documentTitle="GitOps Run" path={[{ label: 'GitOps Run' }]}>
      <ContentWrapper loading={false} errorMessage={undefined}>
        {sessions?.length ? (
          <GitOpsRunTable sessions={sessions} />
        ) : (
          <NoRunsMessage />
        )}
      </ContentWrapper>
    </PageTemplate>
  );
};

export default GitOpsRun;
