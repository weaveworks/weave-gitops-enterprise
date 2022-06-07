import {
  FilterableTable,
  filterConfigForString,
  theme,
} from '@weaveworks/weave-gitops';
import moment from 'moment';
import { FC } from 'react';
import styled, { ThemeProvider } from 'styled-components';
import { usePolicyStyle } from '../../Policies/PolicyStyles';

import { CheckCircle, Error, RemoveCircle } from '@material-ui/icons';
import LinearProgress from '@material-ui/core/LinearProgress';

interface Props {
  canaries: any[];
}
const BorderLinearProgress = styled(LinearProgress)(() => ({
  height: 10,
  width: '100%',
  marginRight: '8px',
  borderRadius: 5,
  [`&.barColorPrimary`]: {
    backgroundColor: '#D8D8D8',
  },
  [`&.bar`]: {
    borderRadius: 5,
    backgroundColor:  '#1a90ff',
  },
}));

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

export const CanaryTable: FC<Props> = ({ canaries }) => {
  const classes = usePolicyStyle();

  const initialFilterState = {
    ...filterConfigForString(canaries, 'clusterName'),
    ...filterConfigForString(canaries, 'namespace'),
  };

  return (
    <div className={classes.root}>
      <ThemeProvider theme={theme}>
        <TableWrapper id="canaries-list">
          <FilterableTable
            key={canaries?.length}
            filters={initialFilterState}
            rows={canaries}
            fields={[
              {
                label: 'Name',
                value: 'name',
              },
              {
                label: 'Status',
                value: ({ status }: any) => {
                  switch (status.phase) {
                    case 'Initialized':
                      return (
                        <div className={classes.flexStart}>
                          <RemoveCircle
                            className={`${classes.severityIcon} ${classes.statusWaiting}`}
                          />
                          Waiting
                        </div>
                      );
                    case 'Succeeded':
                      return (
                        <div className={classes.flexStart}>
                          <CheckCircle
                            className={`${classes.severityIcon} ${classes.statusReady}`}
                          />
                          Ready
                        </div>
                      );
                    case 'Progressing':
                      return (
                        <div className={classes.flexStart}>
                          <BorderLinearProgress
                            variant="determinate"
                            value={status.canaryWeight}
                          />
                          {Math.ceil(status.canaryWeight / 100)} / 10
                        </div>
                      );
                    default:
                      return (
                        <div className={classes.flexStart}>
                          <Error
                            className={`${classes.severityIcon} ${classes.statusFailed}`}
                          />
                          Failed
                        </div>
                      );
                  }
                },
              },
              {
                label: 'Cluster',
                value: 'clusterName',
                textSearchable: true,
              },
              {
                label: 'Namespace',
                value: 'namespace',
              },
              {
                label: 'Target',
                value: (c: any) => c.target_reference.name,
              },
              {
                label: 'Message',
                value: (c: any) => c.status.conditions[0].message,
              },
              {
                label: 'Last Updated',
                value: (c: any) =>
                  moment(c.status.conditions[0].lastUpdateTime).fromNow(),
              },
            ]}
          />
        </TableWrapper>
      </ThemeProvider>
    </div>
  );
};
