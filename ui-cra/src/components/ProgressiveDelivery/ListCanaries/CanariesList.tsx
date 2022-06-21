import { CanaryTable } from './Table';
import {
  ListCanariesResponse,
  ProgressiveDeliveryService,
} from '@weaveworks/progressive-delivery';
import { useQuery } from 'react-query';
import { LoadingPage } from '@weaveworks/weave-gitops';
import { Alert } from '@material-ui/lab';
import { Canary } from '@weaveworks/progressive-delivery/api/prog/types.pb';

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
  return (
    <>
      {isLoading && <LoadingPage />}
      {error && <Alert severity="error">{error.message}</Alert>}
      {data?.canaries?.length && (
        <CanaryTable canaries={data.canaries as Canary[]} />
      )}
    </>
  );
};

export default ProgressiveDelivery;
