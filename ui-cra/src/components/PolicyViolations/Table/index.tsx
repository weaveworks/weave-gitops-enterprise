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
import { PolicyValidation } from '../../../capi-server/capi_server.pb';
import PolicyViolationRow from './Row';
import { PolicyStyles } from '../../Policies/PolicyStyles';

const localMuiTheme = createTheme({
  ...muiTheme,
  shadows: Array(25).fill('none') as Shadows,
});

interface Props {
  violations: PolicyValidation[] | undefined;
}

export const PolicyViolationsTable: FC<Props> = ({ violations }) => {
  const classes = PolicyStyles.useStyles();

  return (
    <div className={`${classes.root}`} id="policies-violations-list">
      <ThemeProvider theme={localMuiTheme}>
        <Paper className={classes.paper}>
          <Table className={classes.table} size="small">
            {violations?.length === 0 ? (
              <caption>No Violations configured</caption>
            ) : null}
            <TableHead className={classes.tableHead}>
              <TableRow>
                <TableCell align="left">
                  <ColumnHeaderTooltip title="Name configured in management UI">
                    <span className={classes.headerCell}>Message</span>
                  </ColumnHeaderTooltip>
                </TableCell>

                <TableCell align="left">
                  <span className={classes.headerCell}>Severity</span>
                </TableCell>
                <TableCell align="left">
                  <span className={classes.headerCell}>Violation Time</span>
                </TableCell>
                <TableCell align="left">
                  <span className={classes.headerCell}>Application</span>
                </TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {violations?.map((violation: PolicyValidation) => {
                return (
                  <PolicyViolationRow
                    violation={violation}
                    key={violation.id}
                  />
                );
              })}
            </TableBody>
          </Table>
        </Paper>
      </ThemeProvider>
    </div>
  );
};
