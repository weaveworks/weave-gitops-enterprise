import { createContext, Dispatch, useContext } from 'react';
import { Profile } from '../../types/custom';

interface ProfilesContext {
  profiles: Profile[] | [];
  loading: boolean;
  error: string | null;
  setError: Dispatch<React.SetStateAction<string | null>>;
  getProfile: (name: string) => Profile | null;
  profilePreview: string | null;
  setProfilePreview: Dispatch<React.SetStateAction<string | null>>;
  renderProfile: (data: any) => void;
  activeProfile: Profile | null;
  setActiveProfile: Dispatch<React.SetStateAction<Profile | null>>;
}

export const Profiles = createContext<ProfilesContext | null>(null);

export default () => useContext(Profiles) as ProfilesContext;
