import { Button, formatURL, Icon, IconType } from '@weaveworks/weave-gitops';
import { Automation, Source } from '@weaveworks/weave-gitops/ui/lib/objects';
import * as React from 'react';
import { Link } from 'react-router-dom';
import styled from 'styled-components';
import { GitOpsSet } from '../../../api/gitopssets/types.pb';
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
  | Pipeline
  | GitOpsSet;

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
    case 'GitOpsSet':
      return formatURL(Routes.EditResource, {
        name: (resource as GitOpsSet)?.name,
        namespace: (resource as GitOpsSet)?.namespace,
        kind: resource.type,
        clusterName: (resource as GitOpsSet)?.clusterName,
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
            startIcon={<Icon type={IconType.SettingsIcon} size="base" />}
            disabled={disabled}
          >
            Edit
          </EditWrapper>
        </div>
      </Tooltip>
    </Link>
  );
};
