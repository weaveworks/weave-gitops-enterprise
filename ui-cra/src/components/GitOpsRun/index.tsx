import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';

import { Routes } from '../../utils/nav';
import NoRunsMessage from './NoRunsMessage';

const GitOpsRun = () => {
  return (
    <PageTemplate
      documentTitle="Run"
      path={[{ label: 'Run', url: Routes.GitOpsRun }]}
    >
      <ContentWrapper loading={false} errorMessage={undefined}>
        <NoRunsMessage />
      </ContentWrapper>
    </PageTemplate>
  );
};

export default GitOpsRun;
