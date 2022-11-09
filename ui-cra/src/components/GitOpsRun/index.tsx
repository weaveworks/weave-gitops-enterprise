import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';

import { Routes } from '../../utils/nav';
import { GitOpsRunTable } from './GitOpsRunTable';
import NoRunsMessage from './NoRunsMessage';

const GitOpsRun = () => {
  const sessions = [
    { name: 'session', cliVersion: '1', portForward: '8000', command: 'do it' },
  ];
  return (
    <PageTemplate
      documentTitle="Run"
      path={[{ label: 'Run', url: Routes.GitOpsRun }]}
    >
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
