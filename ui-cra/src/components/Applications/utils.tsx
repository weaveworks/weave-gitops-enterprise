import { useListAutomations, useListSources } from '@weaveworks/weave-gitops';

export const useApplicationsCount = (): number => {
  const { data: automations } = useListAutomations();

  return automations ? automations?.result?.length : 0;
};

export const useSourcesCount = (): number => {
  const { data: sources } = useListSources();

  return sources ? sources?.result?.length : 0;
};
