import { DataTable, filterConfig } from '@weaveworks/weave-gitops';
import { TableWrapper } from '../../../Shared';
import { useGetWorkspaceServiceAccount } from '../../../../contexts/Workspaces';
import moment from 'moment';
import WorkspaceModal from '../WorkspaceModal';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
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
              value: 'name',
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
            {
              label: '',
              value: ({ manifest }) => (
                <WorkspaceModal
                  content={
                    manifest ? (
                      <SyntaxHighlighter
                        language="yaml"
                        wrapLongLines="pre-wrap"
                        showLineNumbers
                      >
                        {manifest}
                      </SyntaxHighlighter>
                    ) : null
                  }
                  title="Service Accounts Manifest"
                  caption="[some command related to retrieving this yaml]"
                  btnName="View Yaml"
                />
              ),
            },
          ]}
        />
      </TableWrapper>
    </WorkspaceTabsWrapper>
  );
};
