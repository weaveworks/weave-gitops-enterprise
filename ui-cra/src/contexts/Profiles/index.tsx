import { createContext, useContext } from 'react';
import { Profile } from '../../types/custom';

interface ProfilesContext {
  profiles: Profile[] | [];
  loading: boolean;
  getProfile: (name: string) => Profile | null;
  getProfileYaml: (profile: Profile) => Promise<any>;
}

export const Profiles = createContext<ProfilesContext | null>(null);

export default () => useContext(Profiles) as ProfilesContext;
