import { createContext, useContext } from 'react';
import { UpdatedProfile } from '../../types/custom';

interface ProfilesContext {
  loading: boolean;
  isLoading: boolean;
  profiles: UpdatedProfile[];
  getProfileYaml: (name: string, version: string) => Promise<any>;
}

export const Profiles = createContext<ProfilesContext | null>(null);

export default () => useContext(Profiles) as ProfilesContext;
