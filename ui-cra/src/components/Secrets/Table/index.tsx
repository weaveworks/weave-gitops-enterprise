import { FC } from 'react';
import {
  DataTable,
  formatURL,
  KubeStatusIndicator,
  Link,
} from '@weaveworks/weave-gitops';
import { TableWrapper } from '../../Shared';
import { ExternalSecretItem } from '../../../cluster-services/cluster_services.pb';
import moment from 'moment';
import { Routes } from '../../../utils/nav';

interface Props {
  secrets: ExternalSecretItem[];
}

export const SecretsTable: FC<Props> = ({ secrets }) => {
  return (
    <TableWrapper id="secrets-list">
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
                  namespace: s.namespace,
                })}
                data-secret-name={s.externalSecretName}
              >
                {s.externalSecretName}
              </Link>
            ),
            textSearchable: true,
            sortValue: ({ externalSecretName }) => externalSecretName,
          },
          {
            label: 'Status',
            value: ({ status }) => (
              <KubeStatusIndicator
                short
                conditions={[
                  {
                    status: status === 'Ready' ? 'True' : 'False',
                    type: status,
                  },
                ]}
              />
            ),
            sortValue: ({ status }) => status,
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
