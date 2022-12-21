import { FC } from 'react';
import { DataTable, filterConfig, formatURL } from '@weaveworks/weave-gitops';
import { TableWrapper } from '../../Shared';
import { ExternalSecretItem } from '../../../cluster-services/cluster_services.pb';
import { Link } from 'react-router-dom';
import { Routes } from '../../../utils/nav';
import { usePolicyStyle } from '../../Policies/PolicyStyles';
import moment from 'moment';
import Status from '../Status';

interface Props {
  secrets: ExternalSecretItem[];
}

export const SecretsTable: FC<Props> = ({ secrets }) => {
  const classes = usePolicyStyle();

  return (
    <TableWrapper
      id="secrets-list"
      style={{ minHeight: 'calc(100vh - 233px)' }}
    >
      <DataTable
        key={secrets?.length}
        rows={secrets}
        fields={[
          {
            label: 'Name',
            value: (s: ExternalSecretItem) => (
              <Link
                to={formatURL(Routes.SecretDetails, {
                  externalSecretName: s.externalSecretName,
                  clusterName: s.clusterName,
                  namespace:s.namespace
                })}
                className={classes.link}
                data-secret-name={s.externalSecretName}
              >
                {s.externalSecretName}
              </Link>
            ),
            textSearchable: true,
            sortValue: ({ name }) => name,
            maxWidth: 650,
          },
          {
            label: 'Status',
            value: ({status})=> <Status status={status} />,
          },
          {
            label: 'Namespace',
            value: 'namespace',
          },
          {
            label: 'Cluster',
            value: 'clusterName',
          },
          {
            label: 'K8s Secret',
            value: 'secretName',
          },
          {
            label: 'Secret Store',
            value: 'secretStore',
          },
          {
            label: 'Age',
            value: ({ timestamp }) => moment(timestamp).fromNow(),
            sortValue: ({ timestamp }) => {
              const t = timestamp && new Date(timestamp).getTime();
              return t * -1;
            },
          },
        ]}
      />
    </TableWrapper>
  );
};
