import { createContext, Dispatch, useContext } from 'react';
import { Credential } from '../../types/custom';

interface CredentialsContext {
  credentials: Credential[] | undefined;
  loading: boolean;
  error: string | null;
  setError: Dispatch<React.SetStateAction<string | null>>;
  getCredential: (name: string) => Credential | null;
}

export const Credentials = createContext<CredentialsContext | null>(null);

export default () => useContext(Credentials) as CredentialsContext;
