import { useCallback, useContext, useState } from 'react';
import { useQuery } from 'react-query';
import {
  CreatePullRequestRequest,
  ListTemplatesResponse,
} from '../cluster-services/cluster_services.pb';
import useNotifications from '../contexts/Notifications';
import { TemplateEnriched } from '../types/custom';
import { useAPI } from '../contexts/API';

const useTemplates = (
  opts: { enabled: boolean } = {
    enabled: true,
  },
) => {
  const [loading, setLoading] = useState<boolean>(false);
  const { setNotifications } = useNotifications();
  const { enterprise } = useAPI();

  const onError = (error: Error) =>
    setNotifications([{ message: { text: error.message }, severity: 'error' }]);

  const { isLoading, data } = useQuery<ListTemplatesResponse, Error>(
    'templates',
    () => enterprise.ListTemplates({}),
    {
      keepPreviousData: true,
      onError,
      ...opts,
    },
  );
  const templates = data?.templates as TemplateEnriched[] | undefined;

  const getTemplate = (templateName: string) =>
    templates?.find(template => template.name === templateName) || null;

  const renderTemplate = enterprise.RenderTemplate;

  const addResource = useCallback(
    ({ ...data }: CreatePullRequestRequest, token: string | null) => {
      setLoading(true);
      return enterprise
        .CreatePullRequest(data, {
          headers: new Headers({ 'Git-Provider-Token': `token ${token}` }),
        })
        .finally(() => setLoading(false));
    },
    [enterprise],
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
