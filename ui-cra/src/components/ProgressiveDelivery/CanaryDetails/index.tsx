import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../../muiTheme';

import { ContentWrapper } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';
import { SectionHeader } from '../../Layout/SectionHeader';

import { LoadingPage } from '@weaveworks/weave-gitops';
import { Alert } from '@material-ui/lab';
import CanaryDetailsSection from './CanaryDetailsSection';
import { useApplicationsCount } from '../../Applications/utils';
import {
  CanaryParams,
  useCanaryDetails,
  useListCanariesCount,
} from '../../../hooks/progressiveDelivery';

function CanaryDetails({ name, namespace, clusterName }: CanaryParams) {
  const applicationsCount = useApplicationsCount();
  const canariesCount = useListCanariesCount();
  const { error, data, isLoading } = useCanaryDetails({
    name,
    namespace,
    clusterName,
  });

  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate documentTitle="WeGo · Policies">
        <SectionHeader
          className="count-header"
          path={[
            {
              label: 'Applications',
              url: '/applications',
              count: applicationsCount,
            },
            {
              label: 'Delivery',
              url: '/applications/delivery',
              count: canariesCount,
            },
            { label: name },
          ]}
        />
        <ContentWrapper>
          {isLoading && <LoadingPage />}
          {error && <Alert severity="error">{error.message}</Alert>}
          {data?.canary && data?.automation && (
            <CanaryDetailsSection
              canary={data.canary}
              automation={data.automation}
            />
          )}
        </ContentWrapper>
      </PageTemplate>
    </ThemeProvider>
  );
}

export default CanaryDetails;
