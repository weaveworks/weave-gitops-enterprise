import React from 'react';
import { useApplications } from '@weaveworks/weave-gitops';

export const useApplicationsCount = (): number => {
  const [applications, setApplications] = React.useState<any[]>([]);
  const { listApplications } = useApplications();

  const listApplicationsLoaded = Boolean(listApplications);

  React.useEffect(() => {
    listApplications &&
      listApplications().then((res: any[]) => setApplications(res));
    // listApplications is a dynamic function that changes a lot
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [listApplicationsLoaded]);

  return applications.length;
};
