import { useListAutomations, useListSources } from '@weaveworks/weave-gitops';

export const useApplicationsCount = (): number => {
  const { data: automations } = useListAutomations();

  return automations?.result?.length || 0;
};

export const useSourcesCount = (): number => {
  const { data: sources } = useListSources();

  return sources?.result?.length || 0;
};
