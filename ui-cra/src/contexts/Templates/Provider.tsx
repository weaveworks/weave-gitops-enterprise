import React, { FC, useCallback, useEffect, useState } from 'react';
import { Template } from '../../types/custom';
import { request } from '../../utils/request';
import { Templates } from './index';

const TemplatesProvider: FC = ({ children }) => {
  const [loading, setLoading] = useState<boolean>(true);
  const [templates, setTemplates] = useState<Template[]>([]);
  const [activeTemplate, setActiveTemplate] = useState<Template | null>(null);
  const [error, setError] = React.useState<string | null>(null);
  const [PRPreview, setPRPreview] = useState<string | null>(null);
  const [creatingPR, setCreatingPR] = useState<boolean>(false);
  const [PRurl, setPRurl] = useState<string | null>(null);

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
    setCreatingPR(true);
    request('POST', '/v1/pulls', {
      body: JSON.stringify(data),
    })
      .then(data => setPRurl(data.webUrl))
      .catch(err => setError(err.message))
      .finally(() => setCreatingPR(false));
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
        loading,
        activeTemplate,
        setActiveTemplate,
        getTemplate,
        error,
        addCluster,
        renderTemplate,
        PRPreview,
        creatingPR,
        PRurl,
        setPRurl,
      }}
    >
      {children}
    </Templates.Provider>
  );
};

export default TemplatesProvider;
