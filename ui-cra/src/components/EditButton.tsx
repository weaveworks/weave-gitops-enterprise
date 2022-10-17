import * as React from 'react';
import { Link } from 'react-router-dom';
import styled from 'styled-components';
import { Button } from '@weaveworks/weave-gitops';
import { Automation, Source } from '@weaveworks/weave-gitops/ui/lib/objects';
import EditIcon from '@material-ui/icons/Edit';
import { Tooltip } from './Shared';
import { GitopsClusterEnriched } from '../types/custom';
import { getCreateRequestAnnotation } from './Resources/Form/utils';

const EditWrapper = styled(Button)`
  span {
    margin-right: 0px;
  }
`;

export const EditButton: React.FC<{
  resource: GitopsClusterEnriched | Automation | Source;
}> = ({ resource }) => {
  return (
    <Link
      to={{
        pathname: `/resources/${resource.type || 'Cluster'}/${
          resource.name
        }/edit`,
        state: { resource },
      }}
    >
      <Tooltip title={`Edit ${resource.type || 'Cluster'}`} placement="top">
        <div>
          <EditWrapper
            disabled={
              !Boolean(
                resource.type
                  ? getCreateRequestAnnotation(resource)
                  : getCreateRequestAnnotation(resource, 'Cluster'),
              )
            }
            startIcon={<EditIcon fontSize="small" />}
          />
        </div>
      </Tooltip>
    </Link>
  );
};
