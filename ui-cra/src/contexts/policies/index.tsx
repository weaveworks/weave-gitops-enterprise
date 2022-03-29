import { createContext, Dispatch, useContext } from 'react';
import { Policy } from '../../types/custom';

interface PoliciesContext {
  policies:  Policy[];
  loading: boolean;
  error: string | null;
  policy: Policy | null;
  setPolicy: Dispatch<React.SetStateAction<Policy | null>>;
  setError: Dispatch<React.SetStateAction<string | null>>;
  getPolicy: (policyName: string) => void;
}

export const Policies = createContext<PoliciesContext | null>(null);

export default () => useContext(Policies) as PoliciesContext;
