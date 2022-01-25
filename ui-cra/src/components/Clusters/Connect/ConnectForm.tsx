import React, { FC } from 'react';
import styled from 'styled-components';
import { cloneDeep, get, set } from 'lodash';
import { FormState, SetFormState } from '../../../types/form';
import DialogContentText from '@material-ui/core/DialogContentText';
import Typography from '@material-ui/core/Typography';
import Box from '@material-ui/core/Box';
import { theme } from '@weaveworks/weave-gitops';

import { Input } from '../../../utils/form';

const Container = styled.div`
  max-width: 500px;
  margin-right: ${theme.spacing.base};
  margin-left: ${theme.spacing.base};
`;

export const ConnectClusterGeneralForm: FC<{
  connecting: boolean;
  formState: FormState;
  setFormState: SetFormState;
}> = ({ connecting, formState, setFormState }) => {
  const updateFormFn =
    <T extends unknown = HTMLInputElement>(path: (string | number)[]) =>
    (event: React.ChangeEvent<T>) => {
      // @ts-expect-error the typing of event here is broken (previously was typed as `any`) since the typings SelectProps.onChange differ from input.onChange
      const { value } = event.target;
      setFormState(state => set(cloneDeep(state), path, value));
    };

  return (
    <Container>
      <DialogContentText>
        Choose a name to help you identify the cluster you want to connect up.
        You can change this later.
      </DialogContentText>
      <Input
        autoFocus={connecting}
        label="Name"
        onChange={updateFormFn(['cluster', 'name'])}
        value={get(formState, ['cluster', 'name'])}
      />
      {formState.error && (
        <Box ml="125px">
          <Typography color="error">{formState.error}</Typography>
        </Box>
      )}
      <DialogContentText>
        If your cluster has an accessible HTTP endpoint you can provide it here
        and we will link to it.
      </DialogContentText>
      <Input
        label="Ingress URL"
        onChange={updateFormFn(['cluster', 'ingressUrl'])}
        value={get(formState, ['cluster', 'ingressUrl'])}
      />
    </Container>
  );
};
