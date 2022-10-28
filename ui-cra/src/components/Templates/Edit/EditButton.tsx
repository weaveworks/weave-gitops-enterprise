import * as React from 'react';
import { Link } from 'react-router-dom';
import styled from 'styled-components';
import { Button, formatURL } from '@weaveworks/weave-gitops';
import { Automation, Source } from '@weaveworks/weave-gitops/ui/lib/objects';
import EditIcon from '@material-ui/icons/Edit';
import { Tooltip } from '../../Shared';
import { GitopsClusterEnriched } from '../../../types/custom';
import { getCreateRequestAnnotation } from '../Form/utils';
import { Routes } from '../../../utils/nav';

const EditWrapper = styled(Button)`
  span {
    margin-right: 0px;
  }
`;

export const EditButton: React.FC<{
  resource: GitopsClusterEnriched | Automation | Source;
}> = ({ resource }) => {
  const disabled = !Boolean(getCreateRequestAnnotation(resource));

  const link =
    resource.type !== 'GitopsCluster'
      ? formatURL(Routes.EditResource, {
          name: resource.name,
          namespace: resource.namespace,
          kind: resource.type,
          clusterName: (resource as Automation | Source).clusterName,
        })
      : formatURL(Routes.EditResource, {
          name: resource.name,
          namespace: resource.namespace,
          kind: resource.type,
        });

  return (
    <Link to={link} style={{ pointerEvents: disabled ? 'none' : 'all' }}>
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
