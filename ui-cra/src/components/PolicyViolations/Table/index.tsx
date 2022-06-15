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
import { PolicyValidation } from '../../../cluster-services/cluster_services.pb';
import PolicyViolationRow from './Row';
import { usePolicyStyle } from '../../Policies/PolicyStyles';
import {
  FilterableTable,
  filterConfigForString,
  theme,
} from '@weaveworks/weave-gitops/ui';
import styled from 'styled-components';
import { Link } from 'react-router-dom';
import Severity from '../../Policies/Severity';
import moment from 'moment';

const localMuiTheme = createTheme({
  ...muiTheme,
  shadows: Array(25).fill('none') as Shadows,
});

const TableWrapper = styled.div`
  margin-top: ${theme.spacing.medium};
  div[class*='FilterDialog__SlideContainer'],
  div[class*='SearchField'] {
    overflow: hidden;
  }
  div[class*='FilterDialog'] {
    .Mui-checked {
      color: ${theme.colors.primary};
    }
  }
  tr {
    vertical-align:'center')};
  }
  max-width: calc(100vw - 220px);
`;
interface Props {
  violations: PolicyValidation[];
}

export const PolicyViolationsTable: FC<Props> = ({ violations }) => {
  const initialFilterState = {
    ...filterConfigForString(violations, 'clusterName'),
  };
  const classes = usePolicyStyle();
  return (
    <div className={`${classes.root}`} id="policies-violations-list">
      <ThemeProvider theme={localMuiTheme}>
        <TableWrapper>
          <FilterableTable
            key={violations?.length}
            filters={initialFilterState}
            rows={violations}
            fields={[
              {
                label: 'Name configured in management UI',
                value: (v: PolicyValidation) => (
                  <Link
                    to={`/clusters/violations/${v.id}/${v.clusterName}`}
                    className={classes.link}
                  >
                    {v.message}
                  </Link>
                ),
                maxWidth: 650,
              },
              {
                label: 'Severity',
                value: (v: PolicyValidation) => (
                  <Severity severity={v.severity || ''} />
                ),
              },
              {
                label: 'Violation Time',
                value: (v: PolicyValidation) => moment(v.createdAt).fromNow(),
              },
              {
                label: 'Application',
                value: (v: PolicyValidation) => `${v.namespace}/${v.entity}`,
              },
            ]}
          ></FilterableTable>
        </TableWrapper>
        {/* <Paper className={classes.paper}>
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
        </Paper> */}
      </ThemeProvider>
    </div>
  );
};
