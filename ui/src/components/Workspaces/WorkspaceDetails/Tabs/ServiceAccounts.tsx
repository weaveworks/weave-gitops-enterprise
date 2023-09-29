import { DataTable, YamlView } from '@weaveworks/weave-gitops';
import moment from 'moment';
import { useGetWorkspaceServiceAccount } from '../../../../contexts/Workspaces';
import { TableWrapper } from '../../../Shared';
import WorkspaceModal from '../WorkspaceModal';
import WorkspaceTabsWrapper from './WorkspaceTabsWrapper';

export const ServiceAccountsTab = ({
  clusterName,
  workspaceName,
}: {
  clusterName: string;
  workspaceName: string;
}) => {
  const {
    data: serviceAccounts,
    isLoading,
    error,
  } = useGetWorkspaceServiceAccount({
    clusterName,
    workspaceName,
  });

  return (
    <WorkspaceTabsWrapper loading={isLoading} errorMessage={error?.message}>
      <TableWrapper id="service-accounts-list">
        <DataTable
          key={serviceAccounts?.objects?.length}
          rows={serviceAccounts?.objects}
          fields={[
            {
              label: 'Name',
              value: ({ name, namespace, kind, manifest }) => {
                if (manifest) {
                  return (
                    <WorkspaceModal
                      content={
                        <YamlView
                          yaml={manifest}
                          object={{
                            kind: kind,
                            name: name,
                            namespace: namespace,
                          }}
                        />
                      }
                      title="Service Accounts Manifest"
                      btnName={name}
                    />
                  );
                } else return name;
              },
              textSearchable: true,
              maxWidth: 650,
            },
            {
              label: 'Namespace',
              value: 'namespace',
            },
            {
              label: 'Age',
              value: ({ timestamp }) => moment(timestamp).fromNow(),
              sortValue: ({ createdAt }) => {
                const t = createdAt && new Date(createdAt).getTime();
                return t * -1;
              },
            },
          ]}
        />
      </TableWrapper>
    </WorkspaceTabsWrapper>
  );
};
