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
  isLoading: boolean;
}> = ({ resource, isLoading }) => {
  return (
    <Link
      to={{
        pathname: `/resources/${resource.type}/${resource.name}/edit`,
        state: { resource, isLoading },
      }}
    >
      <Tooltip title={`Edit ${resource.type}`} placement="top">
        <div>
          <EditWrapper
            disabled={!Boolean(getCreateRequestAnnotation(resource))}
            startIcon={<EditIcon fontSize="small" />}
          />
        </div>
      </Tooltip>
    </Link>
  );
};
