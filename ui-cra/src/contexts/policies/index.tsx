import { createContext, Dispatch, useContext } from 'react';
import { Policy } from '../../types/custom';

interface PoliciesContext {
  policies:  Policy[];
  loading: boolean;
  error: string | null;
  activePolicy: Policy | null;
  setActivePolicy: Dispatch<React.SetStateAction<Policy | null>>;
  getPolicy: (policyName: string) => Policy | null;
  setError: Dispatch<React.SetStateAction<string | null>>;
}

export const Policies = createContext<PoliciesContext | null>(null);

export default () => useContext(Policies) as PoliciesContext;
