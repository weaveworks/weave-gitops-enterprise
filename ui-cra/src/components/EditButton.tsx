import * as React from 'react';
import { Link } from 'react-router-dom';
import styled from 'styled-components';
import { Button } from '@weaveworks/weave-gitops';
import { Automation, Source } from '@weaveworks/weave-gitops/ui/lib/objects';
import EditIcon from '@material-ui/icons/Edit';
import { Tooltip } from './Shared';

const EditWrapper = styled(Button)`
  span {
    margin-right: 0px;
  }
`;

export const EditButton: React.FC<{
  resource: Automation | Source;
}> = ({ resource }) => {
  const hasCreateRequestAnnotation =
    resource.obj.metadata.annotations?.['templates.weave.works/create-request'];

  return hasCreateRequestAnnotation ? (
    <Link
      to={{
        pathname: `/resources/${resource.name}/edit`,
        state: { resource },
      }}
    >
      <Tooltip title={`Edit ${resource.type}`} placement="top">
        <div>
          <EditWrapper startIcon={<EditIcon fontSize="small" />} />
        </div>
      </Tooltip>
      ;
    </Link>
  ) : null;
};
