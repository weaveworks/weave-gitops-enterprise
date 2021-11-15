import { createContext, useContext } from 'react';
import { UpdatedProfile } from '../../types/custom';

interface ProfilesContext {
  loading: boolean;
  updatedProfiles: UpdatedProfile[];
}

export const Profiles = createContext<ProfilesContext | null>(null);

export default () => useContext(Profiles) as ProfilesContext;
