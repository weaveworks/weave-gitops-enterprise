import { useListAutomations } from '@weaveworks/weave-gitops';

export const useApplicationsCount = (): number => {
  const { data } = useListAutomations(undefined, {})
  return data?.result?.length || 0;
};
