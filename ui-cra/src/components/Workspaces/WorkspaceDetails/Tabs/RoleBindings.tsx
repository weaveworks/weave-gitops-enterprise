import { DataTable } from '@weaveworks/weave-gitops';
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

  return (
    <WorkspaceTabsWrapper loading={isLoading} errorMessage={error?.message}>
      <TableWrapper id="role-bindings-list">
        <DataTable
          key={listRoleBindings?.objects?.length}
          rows={listRoleBindings?.objects}
          fields={[
            {
              label: 'Name',
              value: ({ name, manifest }) => {
                if (manifest) {
                  return (
                    <WorkspaceModal
                      content={
                        <SyntaxHighlighter
                          language="yaml"
                          wrapLongLines="pre-wrap"
                          showLineNumbers
                        >
                          {manifest}
                        </SyntaxHighlighter>
                      }
                      title="RoleBinding Manifest"
                      caption="[some command related to retrieving this yaml]"
                      btnName={name}
                    />
                  );
                } else return name;
              },
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
              sortValue: ({ createdAt }) => {
                const t = createdAt && new Date(createdAt).getTime();
                return t * -1;
              },
            }
          ]}
        />
      </TableWrapper>
    </WorkspaceTabsWrapper>
  );
};
