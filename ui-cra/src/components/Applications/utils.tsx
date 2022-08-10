import { useListAutomations } from '@weaveworks/weave-gitops';

export const useApplicationsCount = (): number => {
  const { data: automations } = useListAutomations();

  return automations?.result?.length || 0;
};
