import { DataTable, Link, Severity, V2Routes, formatURL } from '@weaveworks/weave-gitops';
import moment from 'moment';
import { useGetWorkspacePolicies } from '../../../../contexts/Workspaces';
import { TableWrapper } from '../../../Shared';
import WorkspaceTabsWrapper from './WorkspaceTabsWrapper';

export const PoliciesTab = ({
  clusterName,
  workspaceName,
}: {
  clusterName: string;
  workspaceName: string;
}) => {
  const {
    data: workspacePolicies,
    isLoading,
    error,
  } = useGetWorkspacePolicies({
    clusterName,
    name: workspaceName,
  });

  return (
    <WorkspaceTabsWrapper loading={isLoading} errorMessage={error?.message}>
      <TableWrapper id="workspace-policy-list">
        <DataTable
          key={workspacePolicies?.objects?.length}
          rows={workspacePolicies?.objects}
          fields={[
            {
              label: 'Name',
              value: w => (
                <Link
                  to={formatURL(V2Routes.PolicyDetailsPage, {
                    clusterName: clusterName,
                    id: w.id,
                  })}
                  data-workspace-name={w.name}
                >
                  {w.name}
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
              label: 'Severity',
              value: ({ severity }) => <Severity severity={severity} />,
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
