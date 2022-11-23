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
import { Pipeline } from '../../../api/pipelines/types.pb';
import { GetTerraformObjectResponse } from '../../../api/terraform/terraform.pb';

export type Resource =
  | GitopsClusterEnriched
  | Automation
  | Source
  | GetTerraformObjectResponse
  | Pipeline;

export const getLink = (resource: Resource, type: string) => {
  switch (type) {
    case 'GitopsCluster':
    case 'Pipeline':
      return formatURL(Routes.EditResource, {
        name: (resource as GitopsClusterEnriched | Pipeline).name,
        namespace: (resource as GitopsClusterEnriched | Pipeline).namespace,
        kind: (resource as GitopsClusterEnriched | Pipeline).type,
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
        kind: (resource as Automation | Source).type,
        clusterName: (resource as Automation | Source).clusterName,
      });
    case 'Terraform':
      return formatURL(Routes.EditResource, {
        name: (resource as GetTerraformObjectResponse)?.object?.name,
        namespace: (resource as GetTerraformObjectResponse)?.object?.namespace,
        kind: (resource as GetTerraformObjectResponse)?.object?.type,
        clusterName: (resource as GetTerraformObjectResponse)?.object
          ?.clusterName,
      });
    default:
      return '';
  }
};

export const getType = (resource: Resource) =>
  (resource as GitopsClusterEnriched | Automation | Source | Pipeline).type ||
  (resource as GetTerraformObjectResponse)?.object?.type ||
  '';

const EditWrapper = styled(Button)`
  span {
    margin-right: 0px;
  }
`;

export const EditButton: React.FC<{
  resource: Resource;
  className?: string;
}> = ({ resource, className }) => {
  const disabled = !Boolean(getCreateRequestAnnotation(resource));

  const type = getType(resource);
  const link = getLink(resource, type);

  return (
    <Link to={link} style={{ pointerEvents: disabled ? 'none' : 'all' }}>
      <Tooltip title={`Edit ${type}`} placement="top">
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
