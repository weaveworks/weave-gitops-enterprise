import { ContentWrapper } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';
import { SectionHeader } from '../../Layout/SectionHeader';

import { Alert } from '@material-ui/lab';
import { LoadingPage } from '@weaveworks/weave-gitops';
import { useGetCanaryDetails } from '../../../contexts/ProgressiveDelivery';
import CanaryDetailsSection from './CanaryDetailsSection';

type Props = {
  name: string;
  namespace: string;
  clusterName: string;
};

function CanaryDetails({ name, namespace, clusterName }: Props) {
  const { error, data, isLoading } = useGetCanaryDetails({
    name,
    namespace,
    clusterName,
  });

  return (
    <PageTemplate documentTitle="WeGO · Delivery">
      <SectionHeader
        className="count-header"
        path={[
          {
            label: 'Applications',
            url: '/applications',
          },
          {
            label: 'Delivery',
            url: '/applications/delivery',
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
  );
}

export default CanaryDetails;
