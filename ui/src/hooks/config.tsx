import { useContext } from 'react';
import { useQuery } from 'react-query';
import {
  GetConfigResponse,
} from '../cluster-services/cluster_services.pb';
import { EnterpriseClientContext } from '../contexts/EnterpriseClient';

import { formatError } from '../utils/formatters';
import useNotifications from '../contexts/Notifications';

const useConfig = () => {
  const { setNotifications } = useNotifications();

  const { api } = useContext(EnterpriseClientContext);

  const onError = (error: Error) => setNotifications(formatError(error));

  return useQuery<GetConfigResponse, Error>('config',
    () => api.GetConfig({}),
    {
        onError,
    },
  );
};
export default useConfig;
