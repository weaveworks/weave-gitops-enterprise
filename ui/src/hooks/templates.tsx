import { useCallback, useState } from 'react';
import { useQuery } from 'react-query';
import {
  CreatePullRequestRequest,
  ListTemplatesResponse,
} from '../cluster-services/cluster_services.pb';
import { useEnterpriseClient } from '../contexts/API';
import useNotifications from '../contexts/Notifications';
import { TemplateEnriched } from '../types/custom';

const useTemplates = (
  opts: { enabled: boolean } = {
    enabled: true,
  },
) => {
  const [loading, setLoading] = useState<boolean>(false);
  const { setNotifications } = useNotifications();
  const { clustersService } = useEnterpriseClient();

  const onError = (error: Error) =>
    setNotifications([{ message: { text: error.message }, severity: 'error' }]);

  const { isLoading, data } = useQuery<ListTemplatesResponse, Error>(
    'templates',
    () => clustersService.ListTemplates({}),
    {
      keepPreviousData: true,
      onError,
      ...opts,
    },
  );
  const templates = data?.templates as TemplateEnriched[] | undefined;

  const getTemplate = (templateName: string) =>
    templates?.find(template => template.name === templateName) || null;

  const renderTemplate = clustersService.RenderTemplate;

  const addResource = useCallback(
    ({ ...data }: CreatePullRequestRequest, token: string | null) => {
      setLoading(true);
      return clustersService
        .CreatePullRequest(data, {
          headers: new Headers({ 'Git-Provider-Token': `token ${token}` }),
        })
        .finally(() => setLoading(false));
    },
    [clustersService],
  );

  return {
    isLoading,
    templates,
    loading,
    getTemplate,
    addResource,
    renderTemplate,
  };
};

export default useTemplates;
