import { CanaryTable } from './Table';
import { LoadingPage } from '@weaveworks/weave-gitops';
import { Alert } from '@material-ui/lab';
import { Canary } from '@weaveworks/progressive-delivery/api/prog/types.pb';
import { useListCanaries } from '../../../hooks/progressiveDelivery';


const ProgressiveDelivery = ({
  onCountChange,
}: {
  onCountChange: (count: number) => void;
}) => {
  const { error, data, isLoading } = useListCanaries(res =>
    onCountChange(res.canaries?.length || 0),
  );
  return (
    <>
      {isLoading && <LoadingPage />}
      {error && <Alert severity="error">{error.message}</Alert>}
      {data?.canaries && <CanaryTable canaries={data.canaries as Canary[]} />}
    </>
  );
};

export default ProgressiveDelivery;
