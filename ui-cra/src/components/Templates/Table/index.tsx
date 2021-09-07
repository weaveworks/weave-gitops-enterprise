import React, { FC, useEffect, useState } from 'react';
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
import { ColumnHeaderTooltip } from '../../Shared';
import TemplateRow from './Row';
import { muiTheme } from '../../../muiTheme';
import { ThemeProvider, createTheme } from '@material-ui/core/styles';
import { Shadows } from '@material-ui/core/styles/shadows';
import useTemplates from '../../../contexts/Templates';
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
    table: {
      whiteSpace: 'nowrap',
    },
    tableHead: {
      borderBottom: '1px solid #d8d8d8',
    },
  }),
);

export const TemplatesTable: FC<{ templates: Template[] }> = ({
  templates,
}) => {
  const classes = useStyles();
  const { loading } = useTemplates();
  const [sortedTemplates, setSortedTemplates] = useState<Template[]>(templates);
  const [order, setOrder] = useState<'asc' | 'desc'>('asc');

  const handleClick = () => setOrder(order === 'desc' ? 'asc' : 'desc');

  useEffect(() => {
    const sorted = sortedTemplates.sort();
    const revSorted = sortedTemplates.reverse();
    if (order === 'asc') {
      setSortedTemplates(sorted);
    } else {
      setSortedTemplates(revSorted);
    }
  }, [order, sortedTemplates]);

  return (
    <div id="templates-list">
      <ThemeProvider theme={localMuiTheme}>
        <Paper className={classes.paper}>
          {loading ? (
            <Loader />
          ) : (
            <Table className={classes.table} size="small">
              {sortedTemplates.length === 0 ? (
                <caption>No templates available</caption>
              ) : null}
              <TableHead className={classes.tableHead}>
                <TableRow>
                  <TableCell align="left">
                    <TableSortLabel
                      active={true}
                      direction={order}
                      onClick={handleClick}
                    >
                      <ColumnHeaderTooltip title="Template name">
                        <span>Name</span>
                      </ColumnHeaderTooltip>
                    </TableSortLabel>
                  </TableCell>
                  <TableCell align="left">
                    <ColumnHeaderTooltip title="Template Description">
                      <span>Description</span>
                    </ColumnHeaderTooltip>
                  </TableCell>
                  <TableCell />
                </TableRow>
              </TableHead>
              <TableBody>
                {sortedTemplates.map((template: Template, index: number) => {
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
