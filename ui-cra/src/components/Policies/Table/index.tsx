import {
  Paper,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
} from '@material-ui/core';
import { createStyles, makeStyles } from '@material-ui/styles';
import { FC } from 'react';
import { ColumnHeaderTooltip } from '../../Shared';
import { muiTheme } from '../../../muiTheme';
import { ThemeProvider, createTheme } from '@material-ui/core/styles';
import { Shadows } from '@material-ui/core/styles/shadows';
import { Policy } from '../../../capi-server/capi_server.pb';
import PolicyRow from './Row';
import { PolicyStyles } from '../PolicyStyles';

const localMuiTheme = createTheme({
  ...muiTheme,
  shadows: Array(25).fill('none') as Shadows,
});

interface Props {
  policies: Policy[] | undefined;
}

export const PolicyTable: FC<Props> = ({ policies }) => {
  const classes = PolicyStyles.useStyles();

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
              {policies?.map((policy: Policy) => {
                return <PolicyRow policy={policy} key={policy.name} />;
              })}
            </TableBody>
          </Table>
        </Paper>
      </ThemeProvider>
    </div>
  );
};
