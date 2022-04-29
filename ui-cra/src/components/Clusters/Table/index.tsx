import { Checkbox, Paper } from '@material-ui/core';
import { createStyles, makeStyles } from '@material-ui/styles';
import React, { FC, useEffect } from 'react';
import { Cluster, GitopsCluster } from '../../../types/kubernetes';
import { muiTheme } from '../../../muiTheme';
import { ThemeProvider, createTheme } from '@material-ui/core/styles';
import { Shadows } from '@material-ui/core/styles/shadows';
import useClusters from '../../../contexts/Clusters';
import useNotifications from '../../../contexts/Notifications';
import { useHistory } from 'react-router-dom';
import { Loader } from '../../Loader';
import {
  Field,
  FilterableTable,
  // filterConfigForString,
  theme as weaveTheme,
} from '@weaveworks/weave-gitops';
import _ from 'lodash';

const localMuiTheme = createTheme({
  ...muiTheme,
  shadows: Array(25).fill('none') as Shadows,
});

const useStyles = makeStyles(() =>
  createStyles({
    nameHeaderCell: {
      paddingLeft: weaveTheme.spacing.medium,
    },
    paper: {
      marginBottom: 10,
      marginTop: 10,
      overflowX: 'auto',
      width: '100%',
    },
    root: {
      width: '100%',
    },
    disabled: {
      opacity: 0.5,
    },
    table: {
      whiteSpace: 'nowrap',
    },
    tableHead: {
      borderBottom: `1px solid ${weaveTheme.colors.neutral20}`,
    },
    noMaxWidth: {
      maxWidth: 'none',
    },
    tablePagination: {
      height: '80px',
    },
  }),
);

interface Props {
  filteredClusters: GitopsCluster[] | null;
  count: number | null;
  onEdit: (cluster: Cluster) => void;
}

export const ClustersTable: FC<Props> = ({
  filteredClusters,
  count,
  onEdit,
}) => {
  const classes = useStyles();
  const history = useHistory();
  const { selectedClusters, setSelectedClusters, loading } = useClusters();
  const { notifications } = useNotifications();
  const numSelected = selectedClusters.length;
  const isSelected = (name: string) => selectedClusters.indexOf(name) !== -1;
  const rowCount = filteredClusters?.length || 0;

  const handleSelectAllClick = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (event.target.checked) {
      const newSelected = filteredClusters?.map(cluster => cluster.name);
      setSelectedClusters(newSelected || []);
      return;
    }
    setSelectedClusters([]);
  };

  const handleClick = (event: React.MouseEvent<unknown>, name: string) => {
    const selectedIndex = selectedClusters.indexOf(name);
    let newSelected: string[] = [];

    if (selectedIndex === -1) {
      newSelected = newSelected.concat(selectedClusters, name);
    } else if (selectedIndex === 0) {
      newSelected = newSelected.concat(selectedClusters.slice(1));
    } else if (selectedIndex === selectedClusters.length - 1) {
      newSelected = newSelected.concat(selectedClusters.slice(0, -1));
    } else if (selectedIndex > 0) {
      newSelected = newSelected.concat(
        selectedClusters.slice(0, selectedIndex),
        selectedClusters.slice(selectedIndex + 1),
      );
    }
    setSelectedClusters(newSelected);
  };

  function filterConfigForString(rows: any, key: string) {
    const typeFilterConfig = _.reduce(
      rows,
      (r, v) => {
        const t = v[key];

        if (!_.includes(r, t)) {
          // @ts-ignore
          r.push(t);
        }

        return r;
      },
      [],
    );

    return { [key]: typeFilterConfig };
  }
  const initialFilterState = {
    ...filterConfigForString(filteredClusters, 'name'),
    ...filterConfigForString(filteredClusters, 'namespace'),
  };

  const fields: Field[] = [
    {
      label: 'Name',
      value: 'name',
      sortValue: ({ name }) => name,
      textSearchable: true,
    },
    {
      label: 'Namespace',
      value: 'namespace',
      sortValue: ({ namespace }) => namespace,
      textSearchable: true,
    },
  ];

  useEffect(() => {
    return history.listen(() => {
      setSelectedClusters([]);
    });
  }, [notifications, history, setSelectedClusters]);

  return (
    <div id="clusters-list">
      <ThemeProvider theme={localMuiTheme}>
        <Paper className={classes.paper}>
          {loading ? (
            <Loader />
          ) : (
            <FilterableTable
              fields={fields}
              filters={initialFilterState}
              rows={filteredClusters || []}
            />
          )}
        </Paper>
      </ThemeProvider>
    </div>
  );
};
