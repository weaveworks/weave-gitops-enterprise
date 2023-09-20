import { Query } from '../../api/query/query.pb';
import { createContext } from 'react';

export const QueryServiceContext = createContext<typeof Query>(null as any);

type Props = {
  api: typeof Query;
  children?: any;
};

export default function QueryServiceProvider({ api, children }: Props) {
  return (
    <QueryServiceContext.Provider value={api}>
      {children}
    </QueryServiceContext.Provider>
  );
}
