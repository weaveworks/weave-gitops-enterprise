import { ContentWrapper } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';
import { useGetCanaryDetails } from '../../../contexts/ProgressiveDelivery';
import CanaryDetailsSection from './CanaryDetailsSection';
import { Routes } from '../../../utils/nav';

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
    <PageTemplate
      documentTitle="Delivery"
      path={[
        {
          label: 'Applications',
          url: Routes.Applications,
        },
        {
          label: 'Delivery',
          url: Routes.Canaries,
        },
        { label: name },
      ]}
    >
      <ContentWrapper
        loading={isLoading}
        notification={[
          {
            message: { text: error?.message },
            severity: 'error',
          },
        ]}
      >
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
