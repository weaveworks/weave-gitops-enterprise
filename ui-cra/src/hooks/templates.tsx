import { useCallback, useContext, useState } from 'react';
import { useQuery } from 'react-query';
import { ListTemplatesResponse } from '../cluster-services/cluster_services.pb';
import { TemplateEnriched } from '../types/custom';
import { EnterpriseClientContext } from '../contexts/EnterpriseClient';
import useNotifications from '../contexts/Notifications';

const useTemplates = () => {
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
    },
  );
  const templates = data?.templates as TemplateEnriched[] | undefined;

  const getTemplate = (templateName: string) =>
    templates?.find(template => template.name === templateName) || null;

  const renderTemplate = api.RenderTemplate;

  const addResource: typeof api.CreatePullRequest = useCallback(
    (...args) => {
      setLoading(true);
      return api.CreatePullRequest(...args).finally(() => setLoading(false));
    },
    [api],
  );

  return {
    isLoading,
    templates,
    loading,
    renderTemplate,
    getTemplate,
    addResource,
  };
};

export default useTemplates;
