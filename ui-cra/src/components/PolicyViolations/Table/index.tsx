import { FC } from 'react';
import { muiTheme } from '../../../muiTheme';
import { ThemeProvider, createTheme } from '@material-ui/core/styles';
import { Shadows } from '@material-ui/core/styles/shadows';
import { PolicyValidation } from '../../../cluster-services/cluster_services.pb';
import { usePolicyStyle } from '../../Policies/PolicyStyles';
import { FilterableTable, filterConfig, theme } from '@weaveworks/weave-gitops';
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
    ...filterConfig(violations, 'clusterName'),
    ...filterConfig(violations, 'severity'),
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
                    to={`/clusters/${v.clusterName}/violations/${v.id}`}
                    className={classes.link}
                    data-violation-message={v.message}
                  >
                    {v.message}
                  </Link>
                ),
                textSearchable: true,
                sortValue: ({ message }) => message,
                maxWidth: 650,
              },
              {
                label: 'Severity',
                value: (v: PolicyValidation) => (
                  <Severity severity={v.severity || ''} />
                ),
              },
              {
                label: 'Cluster',
                value: 'clusterName',
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
      </ThemeProvider>
    </div>
  );
};
