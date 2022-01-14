import React, { FC } from 'react';
import styled from 'styled-components';
import { chunk, get } from 'lodash';
import { theme } from '@weaveworks/weave-gitops';
import { useLocalStorage } from 'react-use';
import useAlerts from '../contexts/Alerts';
import { makeStyles } from '@material-ui/core/styles';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableRow from '@material-ui/core/TableRow';
import Paper from '@material-ui/core/Paper';
import moment from 'moment';
import { SectionHeader } from './Layout/SectionHeader';
import { ClusterNameLink, NotAvailable } from './Shared';
import { PageTemplate } from './Layout/PageTemplate';
import { Pagination } from './Pagination';
import { TableFooter } from '@material-ui/core';
import useClusters from '../contexts/Clusters';
import { ContentWrapper, Title } from './Layout/ContentWrapper';
import { Loader } from './Loader';

const alertColor = ({ severity }: { severity: string }) => {
  if (severity === 'warning') {
    return theme.colors.feedbackLight;
  }
  if (severity === 'critical') {
    return theme.colors.alert;
  }
  return theme.colors.success;
};

//
// Shebang
//

const useStyles = makeStyles(t => ({
  table: {
    whiteSpace: 'nowrap',
  },
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

export const AlertsDashboard: FC = () => {
  const classes = useStyles();
  const clustersCount = useClusters().count;
  const { alerts, loading } = useAlerts();
  const [page, setPage] = React.useState<number>(0);
  const [perPage, setPerPage] = useLocalStorage<number>(
    'mccp.alerts.perPage',
    10,
  );

  const pagedAlerts = get(chunk(alerts, perPage), [page]);

  const handleSetPageParams = (page: number, perPage: number) => {
    setPage(page);
    setPerPage(perPage);
  };

  const alertsCount = alerts?.length;

  return (
    <PageTemplate documentTitle="WeGo · Alerts">
      <SectionHeader
        className="count-header"
        path={[
          { label: 'Clusters', url: '', count: clustersCount },
          { label: 'Alerts', url: 'alerts', count: alertsCount },
        ]}
      />
      <ContentWrapper>
        <Title>Alerts dashboard</Title>
        <Paper id="firing-alerts">
          {loading && pagedAlerts?.length === 0 ? (
            <Loader />
          ) : (
            <Table
              className={classes.table}
              size="small"
              aria-label="a dense table"
            >
              {alerts.length === 0 ? <caption>No alerts firing</caption> : null}
              <TableBody>
                {pagedAlerts?.map(alert => (
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
              <TableFooter>
                {alertsCount === 0 ? null : (
                  <TableRow>
                    <Pagination
                      count={alertsCount}
                      onSelectPageParams={handleSetPageParams}
                    />
                  </TableRow>
                )}
              </TableFooter>
            </Table>
          )}
        </Paper>
      </ContentWrapper>
    </PageTemplate>
  );
};
