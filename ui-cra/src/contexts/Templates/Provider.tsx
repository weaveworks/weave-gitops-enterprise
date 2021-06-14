import React, { FC, useCallback, useEffect, useState } from 'react';
import { Template } from '../../types/custom';
import { request } from '../../utils/request';
import { Templates } from './index';

const TemplatesProvider: FC = ({ children }) => {
  const [, setLoading] = useState<boolean>(true);
  const [abortController, setAbortController] =
    useState<AbortController | null>(null);
  const [templates, setTemplates] = useState<Template[]>([]);
  const [activeTemplate, setActiveTemplate] = useState<Template | null>(null);
  const [error, setError] = React.useState<string | null>(null);

  const templatesUrl = '/v1/templates';

  const fetchTemplates = useCallback(() => {
    // abort any inflight requests
    abortController?.abort();

    const newAbortController = new AbortController();
    setAbortController(newAbortController);
    setLoading(true);
    request('GET', templatesUrl, {
      cache: 'no-store',
      signal: newAbortController.signal,
    })
      .then(res => {
        setTemplates(res.templates);
        setError(null);
      })
      .catch(err => {
        if (err.name !== 'AbortError') {
          setError(err.message);
        }
      })
      .finally(() => {
        setLoading(false);
        setAbortController(null);
      });
  }, [abortController]);

  useEffect(() => {
    fetchTemplates();
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  return (
    <Templates.Provider
      value={{
        templates,
        activeTemplate,
        setActiveTemplate,
        error,
      }}
    >
      {children}
    </Templates.Provider>
  );
};

export default TemplatesProvider;
