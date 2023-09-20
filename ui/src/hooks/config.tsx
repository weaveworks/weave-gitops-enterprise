import {
  GetConfigResponse,
} from '../cluster-services/cluster_services.pb';
import { EnterpriseClientContext } from '../contexts/EnterpriseClient';

import useNotifications from '../contexts/Notifications';
import { formatError } from '../utils/formatters';
import { useContext } from 'react';
import { useQuery } from 'react-query';

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
