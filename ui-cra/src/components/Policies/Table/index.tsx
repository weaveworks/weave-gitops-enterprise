import {
  DataTable,
  filterConfig,
  formatURL,
  useFeatureFlags,
} from '@weaveworks/weave-gitops';
import moment from 'moment';
import { FC } from 'react';
import { Link } from 'react-router-dom';
import { Policy } from '../../../cluster-services/cluster_services.pb';
import { Routes } from '../../../utils/nav';
import { TableWrapper } from '../../Shared';
import Mode from '../Mode';
import { usePolicyStyle } from '../PolicyStyles';
import Severity from '../Severity';

interface CustomPolicy extends Policy {
  audit?: string;
  enforce?: string;
}

interface Props {
  policies: CustomPolicy[];
}

export const PolicyTable: FC<Props> = ({ policies }) => {
  const classes = usePolicyStyle();
  const { isFlagEnabled } = useFeatureFlags();

  policies.forEach(policy => {
    policy.audit = policy.modes?.includes('audit') ? 'audit' : '';
    policy.enforce = policy.modes?.includes('admission') ? 'enforce' : '';
  });

  let initialFilterState = {
    ...filterConfig(policies, 'clusterName'),
    ...filterConfig(policies, 'severity'),
    ...filterConfig(policies, 'enforce'),
    ...filterConfig(policies, 'audit'),
  };

  if (isFlagEnabled('WEAVE_GITOPS_FEATURE_TENANCY')) {
    initialFilterState = {
      ...initialFilterState,
      ...filterConfig(policies, 'tenant'),
    };
  }
  return (
    <TableWrapper id="policy-list">
      <DataTable
        key={policies?.length}
        filters={initialFilterState}
        rows={policies}
        fields={[
          {
            label: 'Policy Name',
            value: (p: Policy) => (
              <Link
                to={formatURL(Routes.PolicyDetails, {
                  clusterName: p.clusterName,
                  id: p.id,
                })}
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
            label: 'Audit',
            value: ({ audit }) => <Mode modeName={audit} />,
          },
          {
            label: 'Enforce',
            value: ({ enforce }) => (
              <Mode modeName={enforce ? 'admission' : ''} />
            ),
          },
          ...(isFlagEnabled('WEAVE_GITOPS_FEATURE_TENANCY')
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
            value: ({ createdAt }) => moment(createdAt).fromNow(),
            defaultSort: true,
            sortValue: ({ createdAt }) => {
              const t = createdAt && new Date(createdAt).getTime();
              return t * -1;
            },
          },
        ]}
      />
    </TableWrapper>
  );
};
