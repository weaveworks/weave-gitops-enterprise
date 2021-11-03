import { createContext, useContext } from 'react';
import { Profile, UpdatedProfile } from '../../types/custom';

interface ProfilesContext {
  profiles: Profile[] | [];
  loading: boolean;
  getProfile: (name: string) => Profile | null;
  getProfileYaml: (profile: Profile) => Promise<any>;
  updatedProfiles: UpdatedProfile[];
}

export const Profiles = createContext<ProfilesContext | null>(null);

export default () => useContext(Profiles) as ProfilesContext;
