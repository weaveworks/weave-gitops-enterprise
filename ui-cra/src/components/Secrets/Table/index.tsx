import { FC } from 'react';
import {
  DataTable,
  KubeStatusIndicator,
} from '@weaveworks/weave-gitops';
import { TableWrapper } from '../../Shared';
import { ExternalSecretItem } from '../../../cluster-services/cluster_services.pb';
import moment from 'moment';

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
            value: 'externalSecretName',
            defaultSort: true,
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
            value: ({ timestamp }) => moment.utc(timestamp, 'YYYY-DD-MM HH:MM:SS').fromNow(),
            sortValue: ({ timestamp }) => {
              const t = timestamp && new Date(timestamp).getTime();
              console.log(moment(timestamp, 'YYYY-DD-MM HH:MM:SS').fromNow())
              return t * -1;
            },
          },
        ]}
      />
    </TableWrapper>
  );
};
