import React, { FC, useCallback, useMemo, Dispatch, ChangeEvent } from 'react';
import useCredentials from './../../../../../contexts/Credentials';
import { Credential } from '../../../../../types/custom';
import FormControl from '@material-ui/core/FormControl';
import Select from '@material-ui/core/Select';
import MenuItem from '@material-ui/core/MenuItem';

const Credentials: FC<{
  infraCredential: Credential;
  setInfraCredential: Dispatch<React.SetStateAction<Credential | null>>;
}> = ({ infraCredential, setInfraCredential }) => {
  const { credentials, loading, getCredential } = useCredentials();

  const credentialsItems = useMemo(
    () => [
      ...credentials.map((credential: Credential, index: number) => {
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
    ],
    [credentials],
  );

  const handleSelectCredentials = useCallback(
    (event: ChangeEvent<{ name?: string | undefined; value: unknown }>) => {
      const credential = getCredential(event.target.value as string);
      setInfraCredential(credential);
    },
    [getCredential, setInfraCredential],
  );

  return (
    <div className="credentials">
      <span>Infrastructure provider credentials:</span>
      <FormControl>
        <Select
          disabled={loading}
          id="simple-select-autowidth"
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
