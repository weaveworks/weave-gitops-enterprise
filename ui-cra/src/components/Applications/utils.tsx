import { useListAutomations, useListSources } from '@weaveworks/weave-gitops';

export const useApplicationsCount = (): number => {
  const { data: automations } = useListAutomations(undefined, {});
  return automations?.result?.length || 0;
};

export const useSourcesCount = (): number => {
  const { data: sources } = useListSources(undefined,undefined, {});
  return sources?.result?.length || 0;
};
