import React, { FC, useCallback, useEffect, useState } from 'react';
import { Template } from '../../types/custom';
import { request } from '../../utils/request';
import { Templates } from './index';

const TemplatesProvider: FC = ({ children }) => {
  const [, setLoading] = useState<boolean>(true);
  const [templates, setTemplates] = useState<Template[]>([]);
  const [activeTemplate, setActiveTemplate] = useState<Template | null>(null);
  const [error, setError] = React.useState<string | null>(null);
  const [PRPreview, setPRPreview] = useState<string | null>(null);

  const templatesUrl = '/v1/templates';

  const getTemplate = (templateName: string) =>
    templates.find(template => template.name === templateName) || null;

  const renderTemplate = useCallback(
    ({ ...data }) => {
      setLoading(true);
      request('POST', `${templatesUrl}/${activeTemplate?.name}/render`, {
        body: JSON.stringify({ values: data }),
      })
        .then(data => {
          setPRPreview(data.renderedTemplate);
        })
        .catch(err => setError(err.message))
        .finally(() => {
          setLoading(false);
        });
    },
    [activeTemplate],
  );

  const addCluster = useCallback(({ ...data }) => {
    console.log('addCluster has been called with', data);
    setActiveTemplate(null);
  }, []);

  useEffect(() => {
    setLoading(true);
    request('GET', templatesUrl, {
      cache: 'no-store',
    })
      .then(res => {
        setTemplates(res.templates);
        setError(null);
      })
      .catch(err => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  return (
    <Templates.Provider
      value={{
        templates,
        activeTemplate,
        setActiveTemplate,
        error,
        addCluster,
        renderTemplate,
        PRPreview,
        getTemplate,
      }}
    >
      {children}
    </Templates.Provider>
  );
};

export default TemplatesProvider;
