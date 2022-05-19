import { createContext, useContext } from 'react';

type EnterpriseClientContextType = {
  api: any;
};

export const EnterpriseClientContext =
  createContext<EnterpriseClientContextType | null>(null);

export default () =>
  useContext(EnterpriseClientContext) as EnterpriseClientContextType;
