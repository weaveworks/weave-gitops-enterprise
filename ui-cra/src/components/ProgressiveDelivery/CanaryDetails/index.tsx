import { useParams } from 'react-router-dom';
import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../../muiTheme';

import { ContentWrapper } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';
import { SectionHeader } from '../../Layout/SectionHeader';

import {
  GetCanaryResponse,
  ProgressiveDeliveryService,
} from '@weaveworks/progressive-delivery';
import { useQuery } from 'react-query';
import { LoadingPage } from '@weaveworks/weave-gitops';
import { Alert } from '@material-ui/lab';
import CanaryDetailsSection from './CanaryDetailsSection';
interface CanaryParams {
  name: string;
  namespace: string;
  clusterName: string;
}

function CanaryDetails() {
  const { name, namespace, clusterName } = useParams<CanaryParams>();
  const { error, data, isLoading } = useQuery<GetCanaryResponse, Error>(
    'canaryDetails',
    () =>
      ProgressiveDeliveryService.GetCanary({
        name,
        namespace,
        clusterName,
      }),
    {},
  );

  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate documentTitle="WeGo · Policies">
        <SectionHeader
          className="count-header"
          path={[
            { label: 'Applications', url: '/applications' },
            { label: 'Delivery', url: '/applications/delivery' },
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
