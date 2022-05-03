import {
  Checkbox,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  TableSortLabel,
} from '@material-ui/core';
import { createStyles, makeStyles } from '@material-ui/styles';
import React, { FC, useEffect, useState } from 'react';
import { ColumnHeaderTooltip } from '../../Shared';
import ClusterRow from './Row';
import { muiTheme } from '../../../muiTheme';
import { ThemeProvider, createTheme } from '@material-ui/core/styles';
import { Shadows } from '@material-ui/core/styles/shadows';
import useClusters, { GitopsClusterEnriched } from '../../../contexts/Clusters';
import useNotifications from '../../../contexts/Notifications';
import { useHistory } from 'react-router-dom';
import { Loader } from '../../Loader';
import {
  ChipGroup,
  Flex,
  SearchField,
  State,
  theme as weaveTheme,
  IconButton,
  FilterDialog,
  filterRows,
  filterText,
  initialFormState,
  formStateToFilters,
  Icon,
  IconType,
  toPairs,
} from '@weaveworks/weave-gitops';
// can we do this for multiple fields?
import { Field } from '@weaveworks/weave-gitops/ui/components/DataTable';
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
  filteredClusters: GitopsClusterEnriched[] | null;
  count: number | null;
  disabled?: boolean;
  onSortChange: (order: string) => void;
  order: string;
  orderBy: string;
  fields: Field[];
  filters: { [key: string]: string[] };
  dialogOpen?: boolean;
  onDialogClose?: () => void;
  rows: any[];
}

export const ClustersTable: FC<Props> = ({
  filteredClusters,
  count,
  disabled,
  onSortChange,
  order,
  orderBy,
  fields,
  filters,
  dialogOpen,
  onDialogClose,
  rows,
}) => {
  const classes = useStyles();
  const history = useHistory();
  const { selectedClusters, setSelectedClusters, loading } = useClusters();
  const { notifications } = useNotifications();
  const numSelected = selectedClusters.length;
  const isSelected = (name?: string) =>
    selectedClusters.indexOf(name || '') !== -1;
  const rowCount = filteredClusters?.length || 0;

  const handleSelectAllClick = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (event.target.checked) {
      const newSelected =
        filteredClusters?.map(cluster => cluster.name || '') || [];
      setSelectedClusters(newSelected);
      return;
    }
    setSelectedClusters([]);
  };

  const handleClick = (event: React.MouseEvent<unknown>, name?: string) => {
    const selectedIndex = selectedClusters.indexOf(name || '');
    let newSelected: string[] = [];

    if (selectedIndex === -1) {
      newSelected = newSelected.concat(selectedClusters, name || '');
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

  // * Filtering *

  const [filterDialogOpen, setFilterDialogOpen] = useState(dialogOpen);
  const [filterState, setFilterState] = useState<State>({
    filters,
    formState: initialFormState(filters),
    textFilters: [],
  });
  let filtered = filterRows(rows, filterState.filters);
  filtered = filterText(filtered, fields, filterState.textFilters);
  const chips = toPairs(filterState);

  const handleChipRemove = (chips: string[]) => {
    const next = {
      ...filterState,
    };

    _.each(chips, chip => {
      next.formState[chip] = false;
    });

    const filters = formStateToFilters(next.formState);

    const textFilters = _.filter(next.textFilters, f => !_.includes(chips, f));

    setFilterState({ formState: next.formState, filters, textFilters });
  };

  const handleTextSearchSubmit = (val: string) => {
    setFilterState({
      ...filterState,
      textFilters: _.uniq(_.concat(filterState.textFilters, val)),
    });
  };

  const handleClearAll = () => {
    setFilterState({
      filters: {},
      formState: initialFormState(filters),
      textFilters: [],
    });
  };

  const handleFilterSelect = (filters: any, formState: any) => {
    setFilterState({ ...filterState, filters, formState });
  };

  useEffect(() => {
    return history.listen(() => {
      setSelectedClusters([]);
    });
  }, [notifications, history, setSelectedClusters]);

  return (
    <div
      className={`${classes.root} ${disabled ? classes.disabled : ''}`}
      id="clusters-list"
    >
      <ThemeProvider theme={localMuiTheme}>
        <Paper className={classes.paper}>
          {loading ? (
            <Loader />
          ) : (
            <>
              <Flex wide align>
                <ChipGroup
                  chips={chips}
                  onChipRemove={handleChipRemove}
                  onClearAll={handleClearAll}
                />
                <Flex align wide end>
                  <SearchField onSubmit={handleTextSearchSubmit} />
                  <IconButton
                    onClick={() => setFilterDialogOpen(!filterDialogOpen)}
                    variant={filterDialogOpen ? 'contained' : 'text'}
                    color="inherit"
                  >
                    <Icon
                      type={IconType.FilterIcon}
                      size="medium"
                      color="neutral30"
                    />
                  </IconButton>
                </Flex>
              </Flex>
              <Flex wide tall>
                <Table className={classes.table} size="small">
                  {filteredClusters?.length === 0 ? (
                    <caption>No clusters configured</caption>
                  ) : null}
                  <TableHead className={classes.tableHead}>
                    <TableRow>
                      <TableCell padding="checkbox">
                        <Checkbox
                          indeterminate={
                            numSelected > 0 && numSelected < rowCount
                          }
                          checked={rowCount > 0 && numSelected === rowCount}
                          onChange={handleSelectAllClick}
                          inputProps={{ 'aria-label': 'select all clusters' }}
                          style={{
                            color: weaveTheme.colors.primary,
                          }}
                        />
                      </TableCell>
                      <TableCell
                        className={classes.nameHeaderCell}
                        align="left"
                      >
                        <TableSortLabel
                          disabled={disabled}
                          active={orderBy === 'Name'}
                          direction={
                            orderBy === 'Name'
                              ? (order as 'asc' | 'desc')
                              : 'asc'
                          }
                          onClick={() => onSortChange('Name')}
                        >
                          <ColumnHeaderTooltip title="Name configured in management UI">
                            <span>Name</span>
                          </ColumnHeaderTooltip>
                        </TableSortLabel>
                      </TableCell>
                      <TableCell />
                      <TableCell align="left">
                        <TableSortLabel
                          disabled={disabled}
                          active={orderBy === 'ClusterStatus'}
                          direction={
                            orderBy === 'ClusterStatus'
                              ? (order as 'asc' | 'desc')
                              : 'asc'
                          }
                          onClick={() => onSortChange('ClusterStatus')}
                        >
                          <ColumnHeaderTooltip
                            title={
                              <span>
                                Shows the status of your clusters based on Agent
                                connection and Alertmanager alerts
                              </span>
                            }
                          >
                            <span>Status</span>
                          </ColumnHeaderTooltip>
                        </TableSortLabel>
                      </TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {filteredClusters?.map(
                      (cluster: GitopsClusterEnriched, index: number) => {
                        const isItemSelected = isSelected(cluster?.name);
                        return (
                          <ClusterRow
                            key={cluster.name}
                            index={index}
                            cluster={cluster}
                            aria-checked={isItemSelected}
                            onCheckboxClick={event =>
                              handleClick(event, cluster?.name)
                            }
                            selected={isItemSelected}
                          />
                        );
                      },
                    )}
                  </TableBody>
                </Table>
                <FilterDialog
                  onFilterSelect={handleFilterSelect}
                  filterList={filters}
                  formState={filterState.formState}
                  open={filterDialogOpen}
                />
              </Flex>
            </>
          )}
        </Paper>
      </ThemeProvider>
    </div>
  );
};
