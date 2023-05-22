import { Button } from '@weaveworks/weave-gitops';
import { Workspace } from '../../../cluster-services/cluster_services.pb';
import { generateRowHeaders, SectionRowHeader } from '../../RowHeader';
import FilterListIcon from '@material-ui/icons/FilterList';
import { useNavigate } from 'react-router-dom';
import { useWorkspaceStyle } from '../WorkspaceStyles';
import { toFilterQueryString } from '../../../utils/FilterQueryString';

function WorkspaceHeaderSection({ name, namespaces, clusterName }: Workspace) {
  const classes = useWorkspaceStyle();
  // const history = useHistory();
  const navigate = useNavigate();
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
          navigate(`/applications?filters=${filtersValues}`);
        }}
        className={classes.navigateBtn}
      >
        <FilterListIcon className={classes.filterIcon} />
        go to TENANT applications
      </Button>
      {generateRowHeaders(defaultHeaders)}
    </>
  );
}

export default WorkspaceHeaderSection;
