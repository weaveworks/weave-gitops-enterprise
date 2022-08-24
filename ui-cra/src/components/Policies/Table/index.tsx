import { FC } from 'react';
import { Policy } from '../../../cluster-services/cluster_services.pb';
import { usePolicyStyle } from '../PolicyStyles';
import {
  FilterableTable,
  filterConfig,
  useFeatureFlags,
} from '@weaveworks/weave-gitops';
import { Link } from 'react-router-dom';
import Severity from '../Severity';
import moment from 'moment';
import { TableWrapper } from '../../Shared';

interface Props {
  policies: Policy[];
}

export const PolicyTable: FC<Props> = ({ policies }) => {
  const classes = usePolicyStyle();
  const { data } = useFeatureFlags();
  const flags = data?.flags || {};

  let initialFilterState = {
    ...filterConfig(policies, 'clusterName'),
    ...filterConfig(policies, 'severity'),
  };

  if (flags.WEAVE_GITOPS_FEATURE_TENANCY === 'true') {
    initialFilterState = {
      ...initialFilterState,
      ...filterConfig(policies, 'tenant'),
    };
  }

  return (
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
          ...(flags.WEAVE_GITOPS_FEATURE_TENANCY === 'true'
            ? [{ label: 'Tenant', value: 'tenant' }]
            : []),
          {
            label: 'Severity',
            value: ({ severity }) => <Severity severity={severity || ''} />,
            sortValue: ({ severity }) => severity,
          },
          {
            label: 'Cluster',
            value: 'clusterName',
            sortValue: ({ clusterName }) => clusterName,
          },
          {
            label: 'Age',
            value: (p: Policy) => moment(p.createdAt).fromNow(),
            sortValue: (p: Policy) => {
              const t = p.createdAt && new Date(p.createdAt);
              // Invert the timestamp so the default sort shows the most recent first
              return p.createdAt ? (Number(t) * -1).toString() : '';
            },
          },
        ]}
      />
    </TableWrapper>
  );
};
