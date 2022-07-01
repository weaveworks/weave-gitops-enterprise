import { Alert } from '@material-ui/lab';
import { LoadingPage } from '@weaveworks/weave-gitops';
import { useListCanaries } from '../../../contexts/ProgressiveDelivery';
import { CanaryTable } from './Table';

const ProgressiveDelivery = ({
  onCountChange,
}: {
  onCountChange: (count: number) => void;
}) => {
  const { error, data, isLoading } = useListCanaries();
  return (
    <>
      {isLoading && <LoadingPage />}
      {error && <Alert severity="error">{error.message}</Alert>}
      {data?.canaries?.length && <CanaryTable canaries={data.canaries} />}
    </>
  );
};

export default ProgressiveDelivery;
