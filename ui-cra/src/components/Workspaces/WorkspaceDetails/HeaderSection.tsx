import { Button } from '@weaveworks/weave-gitops';
import { Workspace } from '../../../cluster-services/cluster_services.pb';
import { generateRowHeaders, SectionRowHeader } from '../../RowHeader';
import FilterListIcon from '@material-ui/icons/FilterList';
import { makeStyles } from '@material-ui/core/styles';
import { theme } from '@weaveworks/weave-gitops';
import { useHistory } from 'react-router-dom';

function HeaderSection({ name, namespaces }: Workspace) {
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

  const useStyles = makeStyles({
    btnStyle:{
      marginBottom: theme.spacing.medium,
      marginRight:  theme.spacing.none,
       textTransform: 'uppercase' 
    },
    filterIcon: {
      color: theme.colors.primary10,
      marginRight: theme.spacing.small,
    },
  });
  const classes = useStyles();

  return (
    <>
      <Button
        onClick={() => {
          history.push(`/applications?filters=tenant=${name}`);
        }}
        className={classes.btnStyle}
      >
        <FilterListIcon className={classes.filterIcon} />
        go to TENANT applications
      </Button>
      {generateRowHeaders(defaultHeaders)}
    </>
  );
}

export default HeaderSection;
