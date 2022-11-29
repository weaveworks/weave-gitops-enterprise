import { useContext } from 'react';
import { useQuery } from 'react-query';
import { formatError } from '../../utils/formatters';
import { EnterpriseClientContext } from '../EnterpriseClient';
import useNotifications from './../../contexts/Notifications';

const LIST_WORKSPACES_QUERY_KEY = 'Workspaces';

export function useListListWorkspaces() {
  const { api } = useContext(EnterpriseClientContext);
  const { setNotifications } = useNotifications();
  const onError = (error: Error) => setNotifications(formatError(error));
}