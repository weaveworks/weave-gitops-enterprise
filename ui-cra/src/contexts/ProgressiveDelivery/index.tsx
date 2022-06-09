import * as React from 'react';
import { ProgressiveDeliveryService } from '../../cluster-services/prog.pb';

export type ProgressiveDeliveryContextType = typeof ProgressiveDeliveryService;

export const ProgressiveDeliveryContext =
  React.createContext<ProgressiveDeliveryContextType>(null as any);

type Props = {
  api: ProgressiveDeliveryService;
};

export default function ({ api = ProgressiveDeliveryService, ...rest }: Props) {
  return (
    <ProgressiveDeliveryContext.Provider
      {...rest}
      value={api as typeof ProgressiveDeliveryService}
    />
  );
}

export function useProgressiveDelivery() {
  return React.useContext(ProgressiveDeliveryContext);
}
