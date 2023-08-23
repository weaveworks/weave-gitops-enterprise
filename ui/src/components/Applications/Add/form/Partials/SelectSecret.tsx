import { CircularProgress, IconButton, MenuItem } from '@material-ui/core';
import { Flex, Icon, IconType } from '@weaveworks/weave-gitops';
import _ from 'lodash';
import * as React from 'react';
import { useHistory } from 'react-router-dom';
import styled from 'styled-components';
import { ExternalSecretItem } from '../../../../../cluster-services/cluster_services.pb';
import { useListSecrets } from '../../../../../contexts/Secrets';
import { Select } from '../../../../../utils/form';
import { Routes } from '../../../../../utils/nav';

type Props = {
  className?: string;
  secret: ExternalSecretItem;
  setSecret: React.Dispatch<React.SetStateAction<ExternalSecretItem>>;
};

function SelectSecret({ className, secret, setSecret }: Props) {
  const { data, isLoading } = useListSecrets({});
  const secrets = data?.secrets || [];

  const history = useHistory();

  return (
    <Flex wide align className={className}>
      <Select
        label="SECRET"
        value={secret.secretName || ''}
        onChange={e =>
          setSecret(
            _.find(secrets, secret => secret.secretName === e.target.value) ||
              {},
          )
        }
      >
        <MenuItem value="" key={-1}>
          {isLoading ? <CircularProgress /> : '-'}
        </MenuItem>
        {secrets.map((secret: ExternalSecretItem, index: number) => {
          return (
            <MenuItem value={secret.secretName} key={index}>
              {secret?.secretName}
            </MenuItem>
          );
        })}
      </Select>
      <IconButton onClick={() => history.push(Routes.CreateSecret)}>
        <Icon type={IconType.AddIcon} size="medium" color="neutral30" />
      </IconButton>
    </Flex>
  );
}

export default styled(SelectSecret).attrs({ className: SelectSecret.name })`
  .MuiIconButton-root {
    padding-bottom: 24px;
  }
`;
