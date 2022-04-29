// @ts-nocheck
import React, { FC } from 'react';
import styled from 'styled-components';
import { FormState, SetFormState } from '../../../types/form';
import DialogContentText from '@material-ui/core/DialogContentText';
import { Code, statusBox } from '../../Shared';
import { Poll } from '../../../utils/poll';
import { Cluster } from '../../../types/kubernetes';
import { asMilliseconds } from '../../../utils/time';
import { Loader } from '../../Loader';
import { theme } from '@weaveworks/weave-gitops';

const Container = styled.div`
  margin-right: ${theme.spacing.base};
  margin-left: ${theme.spacing.base};
`;

interface ResponsesById {
  getCluster?: Cluster;
}

export const ConnectClusterConnectionInstructions: FC<{
  formState: FormState;
  setFormState: SetFormState;
  connecting: boolean;
}> = ({ formState, connecting }) => {
  const getCluster = `/gitops/api/clusters/${formState.cluster.id}`;
  const { protocol, host } = window.location;
  // Quoting the URL is important for zsh
  const yamlUrl = `"${protocol}//${host}/gitops/api/agent.yaml?token=${formState.cluster.token}"`;
  return (
    <Container>
      <DialogContentText>To connect your cluster run:</DialogContentText>
      <Code id="instructions">kubectl apply -f {yamlUrl}</Code>
      <Poll<ResponsesById>
        intervalMs={asMilliseconds('5s')}
        queriesById={{ getCluster }}
      >
        {({ responsesById: { getCluster: cluster } }) => {
          if (!cluster) {
            return <Loader />;
          }
          return statusBox(cluster, connecting);
        }}
      </Poll>
    </Container>
  );
};
