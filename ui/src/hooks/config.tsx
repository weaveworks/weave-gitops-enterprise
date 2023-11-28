import { useContext } from 'react';
import { useQuery } from 'react-query';
import { GetConfigResponse } from '../cluster-services/cluster_services.pb';
import { useAPI } from '../contexts/API';
import useNotifications from '../contexts/Notifications';
import { formatError } from '../utils/formatters';

const useConfig = () => {
  const { setNotifications } = useNotifications();

  const { enterprise } = useAPI();

  const onError = (error: Error) => setNotifications(formatError(error));

  return useQuery<GetConfigResponse, Error>(
    'config',
    () => enterprise.GetConfig({}),
    {
      onError,
    },
  );
};
export default useConfig;
