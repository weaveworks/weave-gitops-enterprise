import { Button } from '@weaveworks/weave-gitops';
import { Workspace } from '../../../cluster-services/cluster_services.pb';
import { generateRowHeaders, SectionRowHeader } from '../../RowHeader';
import FilterListIcon from '@material-ui/icons/FilterList';
import { useHistory } from 'react-router-dom';
import { useWorkspaceStyle } from '../WorkspaceStyles';

function HeaderSection({ name, namespaces }: Workspace) {
  const classes = useWorkspaceStyle();
  const history = useHistory();

  const defaultHeaders: Array<SectionRowHeader> = [
    {
      rowkey: 'Tenant Name',
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
          history.push(`/applications?filters=tenant=${name}`);
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

export default HeaderSection;