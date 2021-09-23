import React from 'react';
import { useApplications } from '@weaveworks/weave-gitops';

export const useApplicationsCount = (): number => {
  const [applications, setApplications] = React.useState<any[]>([]);
  const { listApplications } = useApplications();

  React.useEffect(() => {
    listApplications &&
      listApplications().then((res: any[]) => setApplications(res));
  }, [listApplications]);

  return applications.length;
};
