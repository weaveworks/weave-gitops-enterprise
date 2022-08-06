import { FC } from 'react';
import { ThemeProvider, createTheme } from '@material-ui/core/styles';
import { Shadows } from '@material-ui/core/styles/shadows';
import { Policy } from '../../../cluster-services/cluster_services.pb';
import { usePolicyStyle } from '../PolicyStyles';
import { FilterableTable, filterConfig, theme } from '@weaveworks/weave-gitops';
import { Link } from 'react-router-dom';
import Severity from '../Severity';
import moment from 'moment';
import styled from 'styled-components';
import { localEEMuiTheme } from '../../../muiTheme';

const localMuiTheme = createTheme({
  ...localEEMuiTheme,
  shadows: Array(25).fill('none') as Shadows,
});

interface Props {
  policies: Policy[];
}
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
    vertical-align: 'center';
  }
  max-width: calc(100vw - 220px);
`;

export const PolicyTable: FC<Props> = ({ policies }) => {
  const classes = usePolicyStyle();

  const initialFilterState = {
    ...filterConfig(policies, 'clusterName'),
    ...filterConfig(policies, 'severity'),
  };

  return (
    <div className={classes.root}>
      <ThemeProvider theme={localMuiTheme}>
        <TableWrapper id="policy-list">
          <FilterableTable
            key={policies?.length}
            filters={initialFilterState}
            rows={policies}
            fields={[
              {
                label: 'Policy Name',
                value: (p: Policy) => (
                  <Link
                    to={`/policies/details?clusterName=${p.clusterName}&id=${p.id}`}
                    className={classes.link}
                    data-policy-name={p.name}
                  >
                    {p.name}
                  </Link>
                ),
                textSearchable: true,
                sortValue: ({ name }) => name,
                maxWidth: 650,
              },
              {
                label: 'Category',
                value: 'category',
              },
              {
                label: 'Severity',
                value: (p: Policy) => <Severity severity={p.severity || ''} />,
              },
              {
                label: 'Cluster',
                value: 'clusterName',
              },
              {
                label: 'Age',
                value: (p: Policy) => moment(p.createdAt).fromNow(),
              },
            ]}
          />
        </TableWrapper>
      </ThemeProvider>
    </div>
  );
};
