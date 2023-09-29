import { useCallback, useContext, useState } from 'react';
import { useQuery } from 'react-query';
import {
  CreatePullRequestRequest,
  ListTemplatesResponse,
} from '../cluster-services/cluster_services.pb';
import { EnterpriseClientContext } from '../contexts/EnterpriseClient';
import useNotifications from '../contexts/Notifications';
import { TemplateEnriched } from '../types/custom';

const useTemplates = (
  opts: { enabled: boolean } = {
    enabled: true,
  },
) => {
  const [loading, setLoading] = useState<boolean>(false);
  const { setNotifications } = useNotifications();
  const { api } = useContext(EnterpriseClientContext);

  const onError = (error: Error) =>
    setNotifications([{ message: { text: error.message }, severity: 'error' }]);

  const { isLoading, data } = useQuery<ListTemplatesResponse, Error>(
    'templates',
    () => api.ListTemplates({}),
    {
      keepPreviousData: true,
      onError,
      ...opts,
    },
  );
  const templates = data?.templates as TemplateEnriched[] | undefined;

  const getTemplate = (templateName: string) =>
    templates?.find(template => template.name === templateName) || null;

  const renderTemplate = api.RenderTemplate;

  const addResource = useCallback(
    ({ ...data }: CreatePullRequestRequest, token: string | null) => {
      setLoading(true);
      return api
        .CreatePullRequest(data, {
          headers: new Headers({ 'Git-Provider-Token': `token ${token}` }),
        })
        .finally(() => setLoading(false));
    },
    [api],
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
