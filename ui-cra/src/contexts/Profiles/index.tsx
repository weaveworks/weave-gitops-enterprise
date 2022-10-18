import { createContext, Dispatch, useContext } from 'react';
import { UpdatedProfile } from '../../types/custom';

interface ProfilesContext {
  loading: boolean;
  helmRepo: { name: string; namespace: string };
  setHelmRepo: Dispatch<
    React.SetStateAction<{ name: string; namespace: string }>
  >;
  isLoading: boolean;
  profiles: UpdatedProfile[];
  getProfileYaml: (name: string, version: string) => Promise<any>;
}

export const Profiles = createContext<ProfilesContext | null>(null);

export default () => useContext(Profiles) as ProfilesContext;
