import { NotificationsWrapper } from '../../Layout/NotificationsWrapper';
import { useGetCanaryDetails } from '../../../contexts/ProgressiveDelivery';
import CanaryDetailsSection from './CanaryDetailsSection';
import { Routes } from '../../../utils/nav';
import { Page } from '@weaveworks/weave-gitops';

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
    <Page
      loading={isLoading}
      path={[
        {
          label: 'Delivery',
          url: Routes.Canaries,
        },
        { label: name },
      ]}
    >
      <NotificationsWrapper>
        {data?.canary && (
          <CanaryDetailsSection
            canary={data.canary}
            automation={data.automation}
          />
        )}
      </NotificationsWrapper>
    </Page>
  );
}

export default CanaryDetails;
