import { useCallback, useContext, useMemo, useState } from 'react';
import { useQuery } from 'react-query';
import { ListTemplatesResponse } from '../cluster-services/cluster_services.pb';
import { TemplateEnriched } from '../types/custom';
import { request } from '../utils/request';
import { EnterpriseClientContext } from '../contexts/EnterpriseClient';

const useTemplates = () => {
  const [loading, setLoading] = useState<boolean>(false);
  const { api } = useContext(EnterpriseClientContext);

  const { isLoading, data, error } = useQuery<ListTemplatesResponse, Error>(
    'templates',
    () => api.ListTemplates({}),
    {
      keepPreviousData: true,
    },
  );
  const templates = useMemo(
    () =>
      data?.templates?.map(template => ({
        ...template,
        templateType: template?.labels?.['weave.works/template-type'] || '',
      })),
    [data],
  ) as TemplateEnriched[] | undefined;

  const getTemplate = (templateName: string) =>
    templates?.find(template => template.name === templateName) || null;

  const renderTemplate = api.RenderTemplate;

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
    error,
    templates,
    loading,
    getTemplate,
    addCluster,
    renderTemplate,
  };
};

export default useTemplates;
