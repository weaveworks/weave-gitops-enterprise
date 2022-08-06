import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../../muiTheme';

import { ContentWrapper } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';
import { SectionHeader } from '../../Layout/SectionHeader';

import { Alert } from '@material-ui/lab';
import { LoadingPage } from '@weaveworks/weave-gitops';
import {
  useCanariesCount,
  useGetCanaryDetails,
} from '../../../contexts/ProgressiveDelivery';
import { useApplicationsCount } from '../../Applications/utils';
import CanaryDetailsSection from './CanaryDetailsSection';

type Props = {
  name: string;
  namespace: string;
  clusterName: string;
};

function CanaryDetails({ name, namespace, clusterName }: Props) {
  const applicationsCount = useApplicationsCount();
  const canariesCount = useCanariesCount();
  const { error, data, isLoading } = useGetCanaryDetails({
    name,
    namespace,
    clusterName,
  });

  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate documentTitle="WeGo · Delivery">
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
          {data?.canary && (
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
