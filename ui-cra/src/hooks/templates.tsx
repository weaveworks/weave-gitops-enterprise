import { useCallback, useContext, useState } from 'react';
import { useQuery } from 'react-query';
import { ListTemplatesResponse } from '../cluster-services/cluster_services.pb';
import { TemplateEnriched } from '../types/custom';
import { request } from '../utils/request';
import { EnterpriseClientContext } from '../contexts/EnterpriseClient';
import useNotifications from '../contexts/Notifications';

const useTemplates = () => {
  const [loading, setLoading] = useState<boolean>(false);
  const { setNotifications } = useNotifications();
  const { api } = useContext(EnterpriseClientContext);

  const onError = (error: Error) =>
    setNotifications([{ message: { text: error.message }, variant: 'danger' }]);

  const { isLoading, data } = useQuery<ListTemplatesResponse, Error>(
    'templates',
    () => api.ListTemplates({}),
    {
      keepPreviousData: true,
      onError,
    },
  );
  const templates = data?.templates?.map(template => ({
    ...template,
    templateType: template?.labels?.['weave.works/template-type'] || '',
  })) as TemplateEnriched[] | undefined;

  const getTemplate = (templateName: string) =>
    templates?.find(template => template.name === templateName) || null;

  const renderTemplate = useCallback((templateName, data) => {
    const templatesUrl = '/v1/templates';
    return request('POST', `${templatesUrl}/${templateName}/render`, {
      body: JSON.stringify(data),
    });
  }, []);

  const addCluster = useCallback(
    ({ ...data }, token: string, templateKind: string) => {
      setLoading(true);
      return request(
        'POST',
        templateKind === 'GitOpsTemplate'
          ? '/v1/tfcontrollers'
          : '/v1/clusters',
        {
          body: JSON.stringify(data),
          headers: new Headers({ 'Git-Provider-Token': `token ${token}` }),
        },
      ).finally(() => setLoading(false));
    },
    [],
  );

  return {
    isLoading,
    templates,
    loading,
    getTemplate,
    addCluster,
    renderTemplate,
  };
};

export default useTemplates;
