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
      children: (
        <div id="workspace-details-header-namespaces">
          {namespaces?.map((namespace, i) => (
            <span key={namespace}>
              {namespace}
              {namespaces.length !== 1 && i < namespaces.length - 1 && ', '}
            </span>
          ))}
        </div>
      ),
    },
  ];

  return (
    <>
      <Button
        onClick={() => {
          history.push(
            `/applications?filters=tenant%3A%20${name}_clusterName%3A%20${clusterName}_`,
          );
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
