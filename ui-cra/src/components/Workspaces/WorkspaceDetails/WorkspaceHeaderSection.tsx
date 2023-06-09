import { Button, Icon, IconType } from '@weaveworks/weave-gitops';
import { useHistory } from 'react-router-dom';
import { Workspace } from '../../../cluster-services/cluster_services.pb';
import { toFilterQueryString } from '../../../utils/FilterQueryString';
import { SectionRowHeader, generateRowHeaders } from '../../RowHeader';
import { useWorkspaceStyle } from '../WorkspaceStyles';

function WorkspaceHeaderSection({ name, namespaces, clusterName }: Workspace) {
  const classes = useWorkspaceStyle();
  const history = useHistory();
  const defaultHeaders: Array<SectionRowHeader> = [
    {
      rowkey: 'Workspace Name',
      value: name,
    },
    {
      rowkey: 'Namespaces',
      value: namespaces?.join(', '),
    },
  ];

  return (
    <>
      <Button
        onClick={() => {
          const filtersValues = toFilterQueryString([
            { key: 'tenant', value: name || '' },
            { key: 'clusterName', value: clusterName || '' },
          ]);
          history.push(`/applications?filters=${filtersValues}`);
        }}
        className={classes.navigateBtn}
      >
        <Icon type={IconType.FilterIcon} color="primary20" size="small" />
        go to TENANT applications
      </Button>
      {generateRowHeaders(defaultHeaders)}
    </>
  );
}

export default WorkspaceHeaderSection;
