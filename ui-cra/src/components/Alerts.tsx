import React, { FC } from 'react';
import styled from 'styled-components';
import { chunk, get } from 'lodash';
import { spacing } from 'weaveworks-ui-components/lib/theme/selectors';
import theme from 'weaveworks-ui-components/lib/theme';
import { useLocalStorage } from 'react-use';
import { ErrorIcon } from '../assets/img/error-icon';
import useAlerts from '../contexts/Alerts';
import { makeStyles } from '@material-ui/core/styles';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableRow from '@material-ui/core/TableRow';
import Paper from '@material-ui/core/Paper';
import Box from '@material-ui/core/Box';
import Select from '@material-ui/core/Select';
import TablePagination, {
  LabelDisplayedRowsArgs,
} from '@material-ui/core/TablePagination';
import moment from 'moment';
import { Theme } from 'weaveworks-ui-components';
import { Count, SectionHeader } from './Layout/SectionHeader';
import { SelectInputProps } from '@material-ui/core/Select/SelectInput';
import { ClusterNameLink, NotAvailable } from './Shared';
import { PageTemplate } from './Layout/PageTemplate';

const alertColor = ({
  severity,
  theme,
}: {
  severity: string;
  theme: Theme;
}) => {
  if (severity === 'warning') {
    return theme.colors.yellow500;
  }
  if (severity === 'critical') {
    return theme.colors.orange600;
  }
  return theme.colors.green500;
};

//
// Shebang
//

const ErrorWrapper = styled.span`
  margin-left: ${spacing('xs')};
`;

const useStyles = makeStyles(t => ({
  table: {},
  head: {
    backgroundColor: t.palette.common.white,
    color: t.palette.text.secondary,
    borderBottom: 'none',
  },
  clusterNameCell: { borderBottom: 'none', width: '70px' },
  createdCell: {
    borderBottom: 'none',
    width: '100px',
    color: t.palette.text.secondary,
  },
  cellContent: {
    width: 0,
    minWidth: '100%',
    overflow: 'hidden',
    whiteSpace: 'nowrap',
    textOverflow: 'ellipsis',
  },
  perPageInput: {
    fontSize: '14px',
    lineHeight: '14px',
    color: t.palette.text.secondary,
  },
}));

//
// Mixing styled-cmp and mat-ui a bit naughty,
// by default mat-ui takes precedence as they its
// injected into the doc last, "&&" reclaims it by
// "doubling" the className
//
const SeverityCell = styled(TableCell)`
  && {
    border-bottom: none;
    width: 80px;
    font-family: ${theme.fontFamilies.monospace};
    color: ${alertColor};
  }
`;

const DescriptionCell = styled(TableCell)`
  && {
    border-bottom: none;
    color: ${alertColor};
  }
`;

const InlineCount = styled(Count.withComponent('span'))`
  font-size: 14px;
  display: inline-block;
  margin-left: 4px;
`;

const Num = styled.span`
  font-weight: bold;
`;

const AlertHeader = styled(SectionHeader)`
  flex-grow: 0;
`;

const PageMeta = styled.span`
  display: inline-block;
  width: 100px;
  text-align: right;
`;

const labelDisplayedRows = ({ from, to, count }: LabelDisplayedRowsArgs) => (
  <PageMeta>
    <Num>{from}</Num>{' '}
    <>
      to <Num>{to}</Num>
    </>{' '}
    of <InlineCount size="small">{count}</InlineCount>
  </PageMeta>
);

export const AlertsDashboard: FC = () => {
  const classes = useStyles();
  const { alerts, error } = useAlerts();
  const [page, setPage] = React.useState<number>(0);
  const [perPage, setPerPage] = useLocalStorage<number>(
    'mccp.alerts.perPage',
    3,
  );

  const pagedAlerts = get(chunk(alerts, perPage ?? 3), [page]);
  const onChangePage = (
    event: React.MouseEvent<HTMLButtonElement> | null,
    pagerNumber: number,
  ) => {
    setPage(pagerNumber);
  };
  const handleChangeRowsPerPage: SelectInputProps['onChange'] = ev => {
    setPerPage(parseInt(String(ev.target.value), 10));
    setPage(0);
  };

  return (
    <PageTemplate documentTitle="WeGo · Alerts">
      <span id="count-header">
        <SectionHeader title="Alerts" count={alerts.length ?? 0} />
      </span>
      <Paper id="firing-alerts">
        {!alerts || alerts.length === 0 ? (
          <Box color="text.secondary" padding="14px" my={3}>
            <i>No alerts firing</i>
            {error && (
              <ErrorWrapper>
                <ErrorIcon error={error} />
              </ErrorWrapper>
            )}
          </Box>
        ) : (
          <Box py={1} mb={3}>
            <Box
              height="32px"
              display="flex"
              alignItems="center"
              px={2}
              color="text.secondary"
            >
              <AlertHeader size="small" title="Firing alerts" />
              <TablePagination
                component="div"
                count={alerts.length}
                rowsPerPageOptions={[]}
                rowsPerPage={perPage ?? 3}
                labelDisplayedRows={labelDisplayedRows}
                page={page}
                onChangePage={onChangePage}
                backIconButtonProps={{ size: 'small' }}
                nextIconButtonProps={{ size: 'small' }}
              />
              <Box ml="auto" alignItems="center" display="flex" fontSize="14px">
                <Box mr={1}>Per page</Box>
                <Select
                  native
                  value={perPage}
                  onChange={handleChangeRowsPerPage}
                  inputProps={{
                    className: classes.perPageInput,
                  }}
                >
                  <option value={1}>1</option>
                  <option value={3}>3</option>
                  <option value={10}>10</option>
                </Select>
              </Box>
            </Box>
            <Table
              className={classes.table}
              size="small"
              aria-label="a dense table"
            >
              <TableBody>
                {pagedAlerts.map(alert => (
                  <TableRow key={alert.id}>
                    <SeverityCell severity={alert.severity}>
                      {alert.severity ? (
                        alert.severity
                      ) : (
                        <NotAvailable>No severity</NotAvailable>
                      )}
                    </SeverityCell>
                    <DescriptionCell severity={alert.severity}>
                      <div
                        title={alert.annotations.description}
                        className={classes.cellContent}
                      >
                        {alert.labels.alertname}{' '}
                        {alert.annotations.description ||
                          alert.annotations.message}
                      </div>
                    </DescriptionCell>
                    <TableCell className={classes.clusterNameCell}>
                      <ClusterNameLink cluster={alert.cluster} />
                    </TableCell>
                    <TableCell className={classes.createdCell}>
                      {moment(alert.starts_at).fromNow()}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </Box>
        )}
      </Paper>
    </PageTemplate>
  );
};
