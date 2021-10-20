import { createContext, Dispatch, useContext } from 'react';
import { Profile } from '../../types/custom';

interface ProfilesContext {
  profiles: Profile[] | [];
  loading: boolean;
  getProfile: (name: string) => Profile | null;
  profilePreview: string | null;
  setProfilePreview: Dispatch<React.SetStateAction<string | null>>;
  renderProfile: (data: any) => void;
}

export const Profiles = createContext<ProfilesContext | null>(null);

export default () => useContext(Profiles) as ProfilesContext;
