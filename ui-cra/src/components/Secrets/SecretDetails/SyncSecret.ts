import { Dispatch } from 'react';
import { SyncExternalSecretsResponse } from '../../../cluster-services/cluster_services.pb';
import { request } from '../../../utils/request';

export const syncSecret = async (
  payload: any,
  setNotifications: Dispatch<React.SetStateAction<any>>,
  setIsLoading: Dispatch<React.SetStateAction<any>>,
): Promise<SyncExternalSecretsResponse> => {
  setIsLoading(true);
  const updateEvent = await request('POST', `/v1/external-secrets/sync`, {
    body: JSON.stringify(payload),
  }).catch(err => {
    setNotifications([
      {
        message: { text: err.message },
        severity: 'error',
        display: 'top',
      },
    ]);
  });
  setIsLoading(false);
  return updateEvent;
};
