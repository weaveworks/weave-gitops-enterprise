import { createContext, Dispatch, useContext } from 'react';
import { UpdatedProfile } from '../../types/custom';

interface ProfilesContext {
  helmRepo: { name: string; namespace: string };
  setHelmRepo: Dispatch<
    React.SetStateAction<{
      name: string;
      namespace: string;
      clusterName: string;
      clusterNamespace: string;
    }>
  >;
  isLoading: boolean;
  profiles: UpdatedProfile[];
  error: Error | null;
}

export const Profiles = createContext<ProfilesContext | null>(null);

export default () => useContext(Profiles) as ProfilesContext;
