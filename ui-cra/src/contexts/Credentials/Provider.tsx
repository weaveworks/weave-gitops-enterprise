import React, { FC, useCallback, useEffect, useState } from 'react';
import { Credential } from '../../types/custom';
import { request } from '../../utils/request';
import { Credentials } from './index';

const CredentialsProvider: FC = ({ children }) => {
  const [loading, setLoading] = useState<boolean>(true);
  const [credentials, setCredentials] = useState<Credential[]>([]);
  const [error, setError] = React.useState<string | null>(null);

  const credentialsUrl = '/v1/credentials';

  const getCredential = (credentialName: string) =>
    credentials.find(credential => credential.name === credentialName) || null;

  const getCredentials = useCallback(() => {
    setLoading(true);
    request('GET', credentialsUrl, {
      cache: 'no-store',
    })
      .then(res => {
        setCredentials(res.credentials);
        setError(null);
      })
      .catch(err => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    getCredentials();
  }, [getCredentials]);

  return (
    <Credentials.Provider
      value={{
        credentials,
        loading,
        error,
        setError,
        getCredential,
      }}
    >
      {children}
    </Credentials.Provider>
  );
};

export default CredentialsProvider;
