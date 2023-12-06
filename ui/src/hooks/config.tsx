import { useContext } from 'react';
import { useQuery } from 'react-query';
import { GetConfigResponse } from '../cluster-services/cluster_services.pb';
import { useEnterpriseClient } from '../contexts/API';
import useNotifications from '../contexts/Notifications';
import { formatError } from '../utils/formatters';

const useConfig = () => {
  const { setNotifications } = useNotifications();

  const { clustersService } = useEnterpriseClient();

  const onError = (error: Error) => setNotifications(formatError(error));

  return useQuery<GetConfigResponse, Error>(
    'config',
    () => clustersService.GetConfig({}),
    {
      onError,
    },
  );
};
export default useConfig;
