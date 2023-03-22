import { Button } from '@weaveworks/weave-gitops';
import { Workspace } from '../../../cluster-services/cluster_services.pb';
import { generateRowHeaders, SectionRowHeader } from '../../RowHeader';
import FilterListIcon from '@material-ui/icons/FilterList';
import { useHistory } from 'react-router-dom';
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
          const filtersValues = encodeURIComponent(
            `tenant: ${name}_clusterName: ${clusterName}`,
          );
          history.push(`/applications?filters=${filtersValues}`);
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
