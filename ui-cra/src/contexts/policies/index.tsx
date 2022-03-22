import { createContext, Dispatch, useContext } from 'react';
import { Template } from '../../types/custom';

interface PoliciesContext {
  policies:  Template[];
  loading: boolean;
  error: string | null;
  activePolicy: Template | null;
  setActivePolicy: Dispatch<React.SetStateAction<Template | null>>;
  getPolicy: (policyName: string) => Template | null;
  setError: Dispatch<React.SetStateAction<string | null>>;
}

export const Policies = createContext<PoliciesContext | null>(null);

export default () => useContext(Policies) as PoliciesContext;
