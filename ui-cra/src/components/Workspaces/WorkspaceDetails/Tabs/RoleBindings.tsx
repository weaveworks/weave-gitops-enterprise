import { DataTable, filterConfig } from '@weaveworks/weave-gitops';
import { TableWrapper } from '../../../Shared';
import { useGetWorkspaceRoleBinding } from '../../../../contexts/Workspaces';
import moment from 'moment';
import { WorkspaceRoleBindingSubject } from '../../../../cluster-services/cluster_services.pb';
import WorkspaceModal from '../WorkspaceModal';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import WorkspaceTabsWrapper from './WorkspaceTabsWrapper';

export const RoleBindingsTab = ({
  clusterName,
  workspaceName,
}: {
  clusterName: string;
  workspaceName: string;
}) => {
  const {
    data: listRoleBindings,
    isLoading,
    error,
  } = useGetWorkspaceRoleBinding({
    clusterName,
    workspaceName,
  });

  let initialFilterState = {
    ...filterConfig(listRoleBindings?.objects, 'name'),
  };

  return (
    <WorkspaceTabsWrapper loading={isLoading} errorMessage={error?.message}>
      <TableWrapper id="role-bindings-list">
        <DataTable
          key={listRoleBindings?.objects?.length}
          rows={listRoleBindings?.objects}
          filters={initialFilterState}
          fields={[
            {
              label: 'Name',
              value: 'name',
              textSearchable: true,
              maxWidth: 550,
            },
            {
              label: 'Namespace',
              value: 'namespace',
            },
            {
              label: 'Bindings',
              value: ({ subjects }) =>
                subjects
                  .map((item: WorkspaceRoleBindingSubject) => item.name)
                  .join(', '),
            },
            {
              label: 'Role',
              value: ({ role }) => role.name,
            },
            {
              label: 'Age',
              value: ({ timestamp }) => moment(timestamp).fromNow(),
              defaultSort: true,
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
                  btnName="view Yaml"
                />
              ),
            },
          ]}
        />
      </TableWrapper>
    </WorkspaceTabsWrapper>
  );
};
