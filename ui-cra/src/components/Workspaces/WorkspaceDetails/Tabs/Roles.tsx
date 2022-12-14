import { DataTable, filterConfig } from '@weaveworks/weave-gitops';
import { TableWrapper } from '../../../Shared';
import { useGetWorkspaceRoles } from '../../../../contexts/Workspaces';
import moment from 'moment';
import { RulesListWrapper } from '../../WorkspaceStyles';
import { WorkspaceRoleRule } from '../../../../cluster-services/cluster_services.pb';
import WorkspaceModal from '../WorkspaceModal';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import WorkspaceTabsWrapper from './WorkspaceTabsWrapper';

const RulesList = ({ rules }: { rules: WorkspaceRoleRule[] }) => {
  return (
    <RulesListWrapper>
      {rules.map((rule: WorkspaceRoleRule, index: number) => (
        <li key={index}>
          <div>
            <label>Resources:</label>
            <span>{rule.resources?.join(', ')}</span>
          </div>
          <div>
            <label>Verbs:</label>
            <span>{rule.verbs?.join(', ')}</span>
          </div>
          <div>
            <label>Api Groups:</label>
            <span>{rule.groups?.join('.')}</span>
          </div>
        </li>
      ))}
    </RulesListWrapper>
  );
};

export const RolesTab = ({
  clusterName,
  workspaceName,
}: {
  clusterName: string;
  workspaceName: string;
}) => {
  const {
    data: roles,
    isLoading,
    error,
  } = useGetWorkspaceRoles({
    clusterName,
    workspaceName,
  });

  let initialFilterState = {
    ...filterConfig(roles?.objects, 'name'),
  };

  return (
    <WorkspaceTabsWrapper loading={isLoading} errorMessage={error?.message}>
      <TableWrapper id="roles-list">
        <DataTable
          key={roles?.objects?.length}
          rows={roles?.objects}
          filters={initialFilterState}
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
              label: 'Rules',
              value: ({ rules }) => (
                <WorkspaceModal
                  content={rules.length ? <RulesList rules={rules} /> : null}
                  title="Service Accounts Manifest"
                  btnName="View Rules"
                  className="customBackgroundColor"
                />
              ),
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
