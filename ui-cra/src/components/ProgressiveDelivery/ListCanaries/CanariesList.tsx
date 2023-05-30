import { useListCanaries } from '../../../contexts/ProgressiveDelivery';
import { CanaryTable } from './Table';

const ProgressiveDelivery = () => {
  const { data } = useListCanaries();

  return <>{data?.canaries && <CanaryTable canaries={data.canaries} />}</>;
};

export default ProgressiveDelivery;
