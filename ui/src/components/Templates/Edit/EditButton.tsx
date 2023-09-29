import { Automation, Source } from '@weaveworks/weave-gitops/ui/lib/objects';
import EditIcon from '@material-ui/icons/Edit';
import { Button, formatURL } from '@weaveworks/weave-gitops';
import * as React from 'react';
import { Link } from 'react-router-dom';
import styled from 'styled-components';
import { Pipeline } from '../../../api/pipelines/types.pb';
import { GetTerraformObjectResponse } from '../../../api/terraform/terraform.pb';
import { GitopsClusterEnriched } from '../../../types/custom';
import { Routes } from '../../../utils/nav';
import { Tooltip } from '../../Shared';
import { getCreateRequestAnnotation } from '../Form/utils';

export type Resource =
  | GitopsClusterEnriched
  | Automation
  | Source
  | GetTerraformObjectResponse
  | Pipeline;

export const getLink = (resource: Resource) => {
  switch (resource.type) {
    case 'GitopsCluster':
    case 'Pipeline':
      return formatURL(Routes.EditResource, {
        name: (resource as GitopsClusterEnriched | Pipeline).name,
        namespace: (resource as GitopsClusterEnriched | Pipeline).namespace,
        kind: resource.type,
      });
    case 'GitRepository':
    case 'Bucket':
    case 'HelmRepository':
    case 'HelmChart':
    case 'Kustomization':
    case 'HelmRelease':
    case 'OCIRepository':
      return formatURL(Routes.EditResource, {
        name: (resource as Automation | Source).name,
        namespace: (resource as Automation | Source).namespace,
        kind: resource.type,
        clusterName: (resource as Automation | Source).clusterName,
      });
    case 'Terraform':
      return formatURL(Routes.EditResource, {
        name: (resource as GetTerraformObjectResponse)?.object?.name,
        namespace: (resource as GetTerraformObjectResponse)?.object?.namespace,
        kind: resource.type,
        clusterName: (resource as GetTerraformObjectResponse)?.object
          ?.clusterName,
      });
    default:
      return '';
  }
};

const EditWrapper = styled(Button)`
  span {
    margin-right: 0px;
  }
`;

export const EditButton: React.FC<{
  resource: Resource;
  className?: string;
}> = ({ resource, className }) => {
  const disabled = !getCreateRequestAnnotation(resource);
  const link = getLink(resource);

  return (
    <Link to={link} style={{ pointerEvents: disabled ? 'none' : 'all' }}>
      <Tooltip title={`Edit ${resource.type}`} placement="top">
        <div className={className}>
          <EditWrapper
            disabled={disabled}
            startIcon={<EditIcon fontSize="small" />}
          />
        </div>
      </Tooltip>
    </Link>
  );
};
