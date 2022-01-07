import React, {
  FC,
  useCallback,
  useState,
  Dispatch,
  ChangeEvent,
  useMemo,
} from 'react';
import useCredentials from './../../../../../contexts/Credentials';
import { Credential } from '../../../../../types/custom';
import FormControl from '@material-ui/core/FormControl';
import Select from '@material-ui/core/Select';
import MenuItem from '@material-ui/core/MenuItem';

const Credentials: FC<{
  onSelect: Dispatch<React.SetStateAction<Credential | null>>;
}> = ({ onSelect }) => {
  const { credentials, loading, getCredential } = useCredentials();
  const [infraCredential, setInfraCredential] = useState<Credential | null>(
    null,
  );

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
      onSelect(credential);
    },
    [getCredential, onSelect],
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
