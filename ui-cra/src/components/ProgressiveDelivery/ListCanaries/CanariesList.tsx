import { useListCanaries } from '../../../contexts/ProgressiveDelivery';
import { NotificationsWrapper } from '../../Layout/NotificationsWrapper';
import { CanaryTable } from './Table';

const ProgressiveDelivery = () => {
  const { data } = useListCanaries();

  return (
    <NotificationsWrapper>
      {data?.canaries && <CanaryTable canaries={data.canaries} />}
    </NotificationsWrapper>
  );
};

export default ProgressiveDelivery;
