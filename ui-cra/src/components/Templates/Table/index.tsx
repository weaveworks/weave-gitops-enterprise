import React, { FC } from 'react';
import {
  Paper,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  TableSortLabel,
  Theme,
} from '@material-ui/core';
import { createStyles, makeStyles } from '@material-ui/styles';
import { range } from 'lodash';
import { ColumnHeaderTooltip, SkeletonRow } from '../../Shared';
import TemplateRow from './Row';
import { muiTheme } from '../../../muiTheme';
import { ThemeProvider, createTheme } from '@material-ui/core/styles';
import { Shadows } from '@material-ui/core/styles/shadows';
import useTemplates from '../../../contexts/Templates';
import useNotifications from '../../../contexts/Notifications';
import { Loader } from '../../Loader';
import { Template } from '../../../types/custom';

const localMuiTheme = createTheme({
  ...muiTheme,
  shadows: Array(25).fill('none') as Shadows,
});

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    nameHeaderCell: {
      paddingLeft: theme.spacing(4),
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
      borderBottom: '1px solid #d8d8d8',
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
  templates: Template[] | null;
  count: number | null;
  disabled?: boolean;
  onSortChange: (order: string) => void;
  order: string;
  orderBy: string;
}

export const TemplateTable: FC<Props> = ({
  templates,
  count,
  disabled,
  onSortChange,
  order,
  orderBy,
}) => {
  const classes = useStyles();
  const { loading } = useTemplates();
  const { notifications } = useNotifications();

  const showSkeleton = !templates;
  const skeletonRows = range(0, 1).map((id, index) => (
    <SkeletonRow index={index} key={id} />
  ));

  return (
    <div
      className={`${classes.root} ${disabled ? classes.disabled : ''}`}
      id="templates-list"
    >
      <ThemeProvider theme={localMuiTheme}>
        <Paper className={classes.paper}>
          {loading ? (
            <Loader />
          ) : (
            <Table className={classes.table} size="small">
              {templates?.length === 0 ? (
                <caption>No templates available</caption>
              ) : null}
              <TableHead className={classes.tableHead}>
                <TableRow>
                  <TableCell className={classes.nameHeaderCell} align="left">
                    <TableSortLabel
                      disabled={disabled}
                      active={orderBy === 'Name'}
                      direction={
                        orderBy === 'Name' ? (order as 'asc' | 'desc') : 'asc'
                      }
                      onClick={() => onSortChange('Name')}
                    >
                      <ColumnHeaderTooltip title="Template name">
                        <span>Name</span>
                      </ColumnHeaderTooltip>
                    </TableSortLabel>
                  </TableCell>
                  <TableCell />
                  <TableCell align="left">
                    <ColumnHeaderTooltip title="Template Version">
                      <span>Version</span>
                    </ColumnHeaderTooltip>
                  </TableCell>
                  <TableCell align="left">
                    <ColumnHeaderTooltip title="Template Description">
                      <span>Description</span>
                    </ColumnHeaderTooltip>
                  </TableCell>
                  <TableCell />
                  <TableCell />
                </TableRow>
              </TableHead>
              <TableBody>
                {showSkeleton && skeletonRows}
                {!showSkeleton &&
                  templates?.map((template: Template, index: number) => {
                    return (
                      <TemplateRow
                        key={template.name}
                        index={index}
                        template={template}
                      />
                    );
                  })}
              </TableBody>
            </Table>
          )}
        </Paper>
      </ThemeProvider>
    </div>
  );
};
