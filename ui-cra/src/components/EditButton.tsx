import * as React from 'react';
import { Link } from 'react-router-dom';
import styled from 'styled-components';
import { Button, formatURL } from '@weaveworks/weave-gitops';
import { Automation, Source } from '@weaveworks/weave-gitops/ui/lib/objects';
import EditIcon from '@material-ui/icons/Edit';
import { Tooltip } from './Shared';
import { GitopsClusterEnriched } from '../types/custom';
import { getCreateRequestAnnotation } from './Templates/Form/utils';
import { Routes } from './../utils/nav';

const EditWrapper = styled(Button)`
  span {
    margin-right: 0px;
  }
`;

export const EditButton: React.FC<{
  resource: GitopsClusterEnriched | Automation | Source;
  isLoading: boolean;
}> = ({ resource, isLoading }) => {
  const disabled = !Boolean(getCreateRequestAnnotation(resource));
  return (
    <Link
      to={formatURL(Routes.EditResource, {
        name: resource.name,
        namespace: resource.namespace,
        type: resource.type,
      })}
      style={{ pointerEvents: disabled ? 'none' : 'all' }}
    >
      <Tooltip title={`Edit ${resource.type}`} placement="top">
        <div>
          <EditWrapper
            disabled={disabled}
            startIcon={<EditIcon fontSize="small" />}
          />
        </div>
      </Tooltip>
    </Link>
  );
};
