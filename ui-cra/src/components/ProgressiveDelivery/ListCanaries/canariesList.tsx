import { CanaryTable } from './Table';
import { useCallback, useEffect, useState } from 'react';
import LoadingError from '../../LoadingError';
import {
  ListCanariesResponse,
  ProgressiveDeliveryService,
} from '../../../cluster-services/prog.pb';
import { Canary } from '../../../cluster-services/types.pb';

const ProgressiveDelivery = ({
  onCountChange,
}: {
  onCountChange: (count: number) => void;
}) => {
  const [counter, setCounter] = useState<number>(0);

  const fetchCanariesAPI = useCallback(() => {
    console.log(`counter call ${counter}`);
    return ProgressiveDeliveryService.ListCanaries({}).then(res => {
      onCountChange(res.canaries?.length || 0);
      return res;
    });
  }, [counter, onCountChange]);

  useEffect(() => {
    const intervalId = setInterval(() => {
      setCounter(prev => prev + 1);
    }, 60000);
    return () => {
      clearInterval(intervalId);
      setCounter(0);
    };
  }, []);
  return (
    <LoadingError fetchFn={fetchCanariesAPI}>
      {({ value }: { value: ListCanariesResponse }) => (
        <>
          {value.canaries?.length ? (
            <CanaryTable canaries={value.canaries as Canary[]} />
          ) : (
            <p>No data to display</p>
          )}
        </>
      )}
    </LoadingError>
  );
};

export default ProgressiveDelivery;
