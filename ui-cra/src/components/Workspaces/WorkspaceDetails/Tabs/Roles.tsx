import { DataTable, YamlView } from '@weaveworks/weave-gitops';
import { TableWrapper } from '../../../Shared';
import { useGetWorkspaceRoles } from '../../../../contexts/Workspaces';
import moment from 'moment';
import { RulesListWrapper } from '../../WorkspaceStyles';
import { WorkspaceRoleRule } from '../../../../cluster-services/cluster_services.pb';
import WorkspaceModal from '../WorkspaceModal';
import WorkspaceTabsWrapper from './WorkspaceTabsWrapper';

const RulesList = ({ rules }: { rules: WorkspaceRoleRule[] }) => {
  return (
    <RulesListWrapper>
      {rules.length ? (
        rules.map((rule: WorkspaceRoleRule, index: number) => (
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
        ))
      ) : (
        <span>No Data</span>
      )}
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

  return (
    <WorkspaceTabsWrapper loading={isLoading} errorMessage={error?.message}>
      <TableWrapper id="roles-list">
        <DataTable
          key={roles?.objects?.length}
          rows={roles?.objects}
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
                      title="Role Manifest"
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
              label: 'Rules',
              value: ({ rules }) => (
                <WorkspaceModal
                  content={rules.length ? <RulesList rules={rules} /> : null}
                  title="Rules"
                  btnName="View Rules"
                  className="customBackgroundColor"
                  wrapDialogContent
                />
              ),
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
