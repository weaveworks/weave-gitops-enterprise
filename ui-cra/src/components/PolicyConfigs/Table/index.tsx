import {
  DataTable,
  filterConfig,
  Link,
  formatURL,
} from '@weaveworks/weave-gitops';
import moment from 'moment';
import { FC } from 'react';
import { PolicyConfigListItem } from '../../../cluster-services/cluster_services.pb';
import { Routes } from '../../../utils/nav';
import {
  PolicyConfigsTableWrapper,
  usePolicyConfigStyle,
  WarningIcon,
  TotalPolicies
} from '../PolicyConfigStyles';

interface Props {
  PolicyConfigs: PolicyConfigListItem[];
}

export const PolicyConfigsTable: FC<Props> = ({ PolicyConfigs }) => {
  const classes = usePolicyConfigStyle();
  let initialFilterState = {
    ...filterConfig(PolicyConfigs, 'name'),
  };

  return (
    <PolicyConfigsTableWrapper id="policyConfigs-list">
      <DataTable
        key={PolicyConfigs?.length}
        filters={initialFilterState}
        rows={PolicyConfigs}
        fields={[
          {
            label: '',
            value: ({ status, clusterName, name }) =>
              status === 'Warning' ? (
                <span
                  title={`One or more policies are not found in cluster ${clusterName}.`}
                  data-testid={`warning-icon-${name}`}
                >
                  <WarningIcon />
                </span>
              ) : (
                ' '
              ),
            maxWidth: 50,
          },
          {
            label: 'Name',
            value: (s: PolicyConfigListItem) => (
              <Link
                to={formatURL(Routes.PolicyConfigsDetails, {
                  clusterName: s.clusterName,
                  name: s.name,
                })}
                data-policyconfig-name={s.name}
              >
                {s.name}
              </Link>
            ),
            textSearchable: true,
            defaultSort: true,
            sortValue: ({ name }) => name,
            maxWidth: 650,
          },
          {
            label: 'Cluster',
            value: 'clusterName',
          },
          {
            label: 'Policy Count',
            sortValue: ({ totalPolicies }) => totalPolicies,
            value: ({ totalPolicies }) => (
              <TotalPolicies wide center>{totalPolicies}</TotalPolicies>
            ),
            maxWidth: 100,
          },
          {
            label: 'Applied To',
            sortValue: ({ match }) => match,
            value: ({ match }) => (
              <span className={classes.capitlize}>{match}</span>
            ),
          },
          {
            label: 'Age',
            value: ({ age }) => moment(age).fromNow(),
            sortValue: ({ age }) => {
              const t = age && new Date(age).getTime();
              return t * -1;
            },
          },
        ]}
      />
    </PolicyConfigsTableWrapper>
  );
};
