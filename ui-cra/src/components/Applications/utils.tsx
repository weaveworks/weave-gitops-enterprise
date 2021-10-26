import React from 'react';
import { useApplications } from '@weaveworks/weave-gitops';
import { Application } from '@weaveworks/weave-gitops/ui/lib/api/applications/applications.pb';

export const useApplicationsCount = (): number => {
  const [applications, setApplications] = React.useState<Application[] | void>(
    [],
  );
  const { listApplications } = useApplications();

  const listApplicationsLoaded = Boolean(listApplications);

  React.useEffect(() => {
    listApplications && listApplications().then(res => setApplications(res));
    // listApplications is a dynamic function that changes a lot
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [listApplicationsLoaded]);

  return applications ? applications.length : 0;
};
