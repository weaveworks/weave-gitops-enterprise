import React from 'react';
import { useApplications } from '@weaveworks/weave-gitops';

export const useApplicationsCount = (): number => {
  const [applications, setApplications] = React.useState<any[]>([]);
  const { listApplications } = useApplications();

  React.useEffect(() => {
    listApplications().then((res: any[]) => setApplications(res));
    // listApplications is a dynamic function that changes a lot
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return applications.length;
};
