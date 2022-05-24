import {
  Paper,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
} from '@material-ui/core';
import { FC } from 'react';
import { ColumnHeaderTooltip } from '../../Shared';
import { muiTheme } from '../../../muiTheme';
import { ThemeProvider, createTheme } from '@material-ui/core/styles';
import { Shadows } from '@material-ui/core/styles/shadows';
import { Policy } from '../../../cluster-services/cluster_services.pb';
import PolicyRow from './Row';
import { usePolicyStyle } from '../PolicyStyles';

const localMuiTheme = createTheme({
  ...muiTheme,
  shadows: Array(25).fill('none') as Shadows,
});

interface Props {
  policies: Policy[] | undefined;
}

export const PolicyTable: FC<Props> = ({ policies }) => {
  const classes = usePolicyStyle();

  return (
    <div className={`${classes.root}`} id="policies-list">
      <ThemeProvider theme={localMuiTheme}>
        <Paper className={classes.paper}>
          <Table className={classes.table} size="small">
            {policies?.length === 0 ? (
              <caption>No policies configured</caption>
            ) : null}
            <TableHead className={classes.tableHead}>
              <TableRow>
                <TableCell align="left">
                  <ColumnHeaderTooltip title="Name configured in management UI">
                    <span className={classes.headerCell}>Policy Name</span>
                  </ColumnHeaderTooltip>
                </TableCell>
                <TableCell align="left">
                  <span className={classes.headerCell}>Cluster</span>
                </TableCell>
                <TableCell align="left">
                  <span className={classes.headerCell}>Category</span>
                </TableCell>
                <TableCell align="left">
                  <span className={classes.headerCell}>Severity</span>
                </TableCell>
                <TableCell align="left">
                  <span className={classes.headerCell}>Age</span>
                </TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {policies?.map((policy: Policy, index) => {
                return <PolicyRow policy={policy} key={index} />;
              })}
            </TableBody>
          </Table>
        </Paper>
      </ThemeProvider>
    </div>
  );
};
