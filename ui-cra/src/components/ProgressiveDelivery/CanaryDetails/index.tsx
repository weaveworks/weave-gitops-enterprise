import { ContentWrapper } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';
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
    <PageTemplate
      documentTitle="Delivery"
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
    >
      <ContentWrapper loading={isLoading} errorMessage={error?.message}>
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
