import { createContext, useContext } from 'react';

export interface VersionData {
  ui: string;
  capiServer?: string;
}

type VersionsContext = {
  versions: VersionData | null;
  entitlement: string | null;
  repositoryURL: string;
};

export const Versions = createContext<VersionsContext | null>(null);

export default () => useContext(Versions) as VersionsContext;
