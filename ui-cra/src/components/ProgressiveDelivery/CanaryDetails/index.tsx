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
  const { data, isLoading } = useGetCanaryDetails({
    name,
    namespace,
    clusterName,
  });

  return (
    <PageTemplate
      documentTitle="Delivery"
      path={[
        {
          label: 'Delivery',
          url: Routes.Canaries,
        },
        { label: name },
      ]}
    >
      <ContentWrapper loading={isLoading}>
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
