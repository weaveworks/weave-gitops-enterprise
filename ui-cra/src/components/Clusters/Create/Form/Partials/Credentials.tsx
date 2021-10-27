import React, { FC, FormEvent, useCallback, useMemo, useState } from 'react';
import useCredentials from './../../../../../contexts/Credentials';
import { Credential } from '../../../../../types/custom';
import { Dropdown, DropdownItem } from 'weaveworks-ui-components';

const Credentials: FC<{ onSelect: any }> = ({ onSelect }) => {
  const { credentials, loading, getCredential } = useCredentials();
  const [infraCredential, setInfraCredential] =
    useState<Credential | null>(null);

  const credentialsItems: DropdownItem[] = useMemo(
    () => [
      ...credentials.map((credential: Credential) => {
        const { kind, namespace, name } = credential;
        return {
          label: `${kind}/${namespace || 'default'}/${name}`,
          value: name || '',
        };
      }),
      { label: 'None', value: '' },
    ],
    [credentials],
  );

  const handleSelectCredentials = useCallback(
    (event: FormEvent<HTMLInputElement>, value: string) => {
      const credential = getCredential(value);
      setInfraCredential(credential);
      onSelect(credential);
    },
    [getCredential, onSelect],
  );

  return (
    <div className="credentials">
      <span>Infrastructure provider credentials:</span>
      <Dropdown
        value={infraCredential?.name}
        disabled={loading}
        items={credentialsItems}
        onChange={handleSelectCredentials}
      />
    </div>
  );
};

export default Credentials;
