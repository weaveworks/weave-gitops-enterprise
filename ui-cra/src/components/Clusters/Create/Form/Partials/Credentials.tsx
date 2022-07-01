import React, { FC, useCallback, Dispatch, ChangeEvent, useMemo } from 'react';
import { Credential } from '../../../../../types/custom';
import FormControl from '@material-ui/core/FormControl';
import Select from '@material-ui/core/Select';
import MenuItem from '@material-ui/core/MenuItem';
import { useListCredentials } from '../../../../../hooks/credentials';

const Credentials: FC<{
  infraCredential: Credential | null;
  setInfraCredential: Dispatch<React.SetStateAction<Credential | null>>;
}> = ({ infraCredential, setInfraCredential }) => {
  const { data, isLoading } = useListCredentials();
  const credentials = useMemo(
    () => data?.credentials || [],
    [data?.credentials],
  );

  const credentialsItems = [
    ...credentials.map((credential: Credential) => {
      const { kind, namespace, name } = credential;
      return (
        <MenuItem key={name} value={name || ''}>
          {`${kind}/${namespace || 'default'}/${name}`}
        </MenuItem>
      );
    }),
    <MenuItem key="None" value="None">
      <em>None</em>
    </MenuItem>,
  ];

  const handleSelectCredentials = useCallback(
    (event: ChangeEvent<{ name?: string | undefined; value: unknown }>) => {
      const credential =
        credentials?.find(
          credential => credential.name === event.target.value,
        ) || null;

      setInfraCredential(credential);
    },
    [credentials, setInfraCredential],
  );

  return (
    <div className="credentials">
      <span>Infrastructure provider credentials:</span>
      <FormControl>
        <Select
          disabled={isLoading}
          value={infraCredential?.name || 'None'}
          onChange={handleSelectCredentials}
          autoWidth
          label="Credentials"
        >
          {credentialsItems}
        </Select>
      </FormControl>
    </div>
  );
};

export default Credentials;
