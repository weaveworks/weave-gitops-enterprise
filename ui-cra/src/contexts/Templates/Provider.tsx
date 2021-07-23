import React, { FC, useCallback, useEffect, useState } from 'react';
import { Template } from '../../types/custom';
import { request } from '../../utils/request';
import { Templates } from './index';
import { useHistory } from 'react-router-dom';

const TemplatesProvider: FC = ({ children }) => {
  const [loading, setLoading] = useState<boolean>(true);
  const [templates, setTemplates] = useState<Template[]>([]);
  const [activeTemplate, setActiveTemplate] = useState<Template | null>(null);
  const [error, setError] = React.useState<string | null>(null);
  const [PRPreview, setPRPreview] = useState<string | null>(null);
  const [creatingPR, setCreatingPR] = useState<boolean>(false);

  const history = useHistory();

  const templatesUrl = '/v1/templates';

  const getTemplate = (templateName: string) =>
    templates.find(template => template.name === templateName) || null;

  const renderTemplate = useCallback(
    data => {
      setLoading(true);
      request('POST', `${templatesUrl}/${activeTemplate?.name}/render`, {
        body: JSON.stringify(data),
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

  const addCluster = useCallback(
    ({ ...data }) => {
      setCreatingPR(true);
      request('POST', '/v1/pulls', {
        body: JSON.stringify(data),
      })
        .then(() => history.push('/clusters'))
        .catch(err => setError(err.message))
        .finally(() => setCreatingPR(false));
    },
    [history],
  );

  const getTemplates = useCallback(() => {
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

  useEffect(() => {
    getTemplates();
    return history.listen(getTemplates);
  }, [history, getTemplates]);

  return (
    <Templates.Provider
      value={{
        templates,
        loading,
        activeTemplate,
        setActiveTemplate,
        getTemplate,
        error,
        setError,
        addCluster,
        renderTemplate,
        PRPreview,
        setPRPreview,
        creatingPR,
      }}
    >
      {children}
    </Templates.Provider>
  );
};

export default TemplatesProvider;
