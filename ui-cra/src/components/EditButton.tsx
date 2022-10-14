import * as React from 'react';
import { Link } from 'react-router-dom';
import styled from 'styled-components';
import { Button } from '@weaveworks/weave-gitops';
import { Automation, Source } from '@weaveworks/weave-gitops/ui/lib/objects';
import EditIcon from '@material-ui/icons/Edit';
import { Tooltip } from './Shared';
import { getCreateRequestAnnotation } from './Clusters/Form/utils';
import { GitopsClusterEnriched } from '../types/custom';

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
        pathname: `/resources/${resource.name}/edit`,
        state: { resource },
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
      ;
    </Link>
  );
};
