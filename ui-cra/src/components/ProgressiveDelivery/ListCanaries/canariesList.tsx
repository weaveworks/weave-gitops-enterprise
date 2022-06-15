import { CanaryTable } from './Table';
import {
  ListCanariesResponse,
  ProgressiveDeliveryService,
} from '../../../cluster-services/prog.pb';
import { Canary } from '../../../cluster-services/types.pb';
import { useQuery } from 'react-query';
import { LoadingPage } from '@weaveworks/weave-gitops';
import { Alert } from '@material-ui/lab';

const CANARIES_POLL_INTERVAL = 60000;

const ProgressiveDelivery = ({
  onCountChange,
}: {
  onCountChange: (count: number) => void;
}) => {
  const { error, data, isLoading } = useQuery<ListCanariesResponse, Error>(
    'canaries',
    () =>
      ProgressiveDeliveryService.ListCanaries({}).then(res => {
        onCountChange(res.canaries?.length || 0);
        return res;
      }),
    {
      refetchInterval: CANARIES_POLL_INTERVAL,
    },
  );

  if (isLoading) {
    return <LoadingPage />;
  }
  return (
    <>
      {error && <Alert severity="error">{error.message}</Alert>}
      {data?.canaries?.length ? (
        <CanaryTable canaries={data.canaries as Canary[]} />
      ) : (
        <p>No Data to display</p>
      )}
    </>
  );
};

export default ProgressiveDelivery;
