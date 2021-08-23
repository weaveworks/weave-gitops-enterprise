import { createContext, useContext } from 'react';
import { Alert } from '../../types/kubernetes';

interface AlertsContext {
  alerts: Alert[] | [];
  loading: boolean;
}

export const Alerts = createContext<AlertsContext | null>(null);

export default () => useContext(Alerts) as AlertsContext;
