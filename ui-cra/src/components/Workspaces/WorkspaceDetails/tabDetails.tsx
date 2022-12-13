import React, { useState } from 'react';
import { Box } from '@material-ui/core';
import {
  Button,
  DataTable,
  filterConfig,
  formatURL,
} from '@weaveworks/weave-gitops';
import { TableWrapper } from '../../Shared';
import { Link } from 'react-router-dom';
import {
  WorkspaceRoleBindingSubject,
  WorkspaceRoleRule,
} from '../../../cluster-services/cluster_services.pb';
import {
  useGetWorkspaceRoles,
  useGetWorkspaceRoleBinding,
  useGetWorkspaceServiceAccount,
  useGetWorkspacePolicies,
} from '../../../contexts/Workspaces';
import { Routes } from '../../../utils/nav';
import {
  WorkspacesTabs,
  WorkspaceTab,
  ViewYamlBtn,
} from '../WorkspaceStyles';
import moment from 'moment';
import WorkspaceModal from './workspaceModal';
import Severity from '../../Policies/Severity';

interface TabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
  empty?: boolean;
}
interface TabContent {
  tabName: string;
  value?: string;
  data?: any;
}

function TabPanel(props: TabPanelProps) {
  const { children, value, index, empty, ...other } = props;
  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`tabpanel-${index}`}
      aria-labelledby={`tab-${index}`}
      {...other}
    >
      {value === index && <Box>{children}</Box>}
    </div>
  );
}

const TabDetails = ({
  clusterName,
  workspaceName,
}: {
  clusterName: string;
  workspaceName: string;
}) => {
  const [selectedTab, setSelectedTab] = useState<number>(0);
  const [dialogContent, setDialogContent] = useState<
    string | WorkspaceRoleRule[]
  >();
  const [isDialogOpen, setIsDialogOpen] = useState<boolean>(false);
  const [contentType, setContentType] = useState<string>('');
  const [dialogTitle, setDialogTitle] = useState<string>('');

  const { data: roles } = useGetWorkspaceRoles({
    clusterName,
    workspaceName,
  });

  const { data: listRoleBindings } = useGetWorkspaceRoleBinding({
    clusterName,
    workspaceName,
  });

  const { data: serviceAccounts } = useGetWorkspaceServiceAccount({
    clusterName,
    workspaceName,
  });
  const { data: workspacePolicies } = useGetWorkspacePolicies({
    clusterName,
    workspaceName,
  });

  const handleChange = (event: React.ChangeEvent<{}>, newValue: number) => {
    setSelectedTab(newValue);
  };
  const viewDialog = (
    type: string,
    title: string,
    content: string | WorkspaceRoleRule[],
  ) => {
    setContentType(type);
    setDialogTitle(title);
    setDialogContent(content);
    setIsDialogOpen(true);
  };

  const generateFilters = (T: any) => {
    return {
      ...filterConfig(T, 'name'),
    };
  };

  const tabsContent: Array<TabContent> = [
    {
      tabName: 'SERVICE ACCOUNTS',
      data: (
        <TableWrapper id="service-accounts-list">
          <DataTable
            key={serviceAccounts?.objects?.length}
            rows={serviceAccounts?.objects}
            filters={generateFilters(serviceAccounts?.objects)}
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
                  <ViewYamlBtn>
                    <Button
                      onClick={() =>
                        viewDialog(
                          'yaml',
                          'Service Accounts Manifest',
                          manifest,
                        )
                      }
                      style={{ marginRight: 0, textTransform: 'uppercase' }}
                    >
                      View Yaml
                    </Button>
                  </ViewYamlBtn>
                ),
              },
            ]}
          />
        </TableWrapper>
      ),
    },
    {
      tabName: 'ROLES',
      data: (
        <TableWrapper id="roles-list">
          <DataTable
            key={roles?.objects?.length}
            rows={roles?.objects}
            filters={generateFilters(roles?.objects)}
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
                  <Button
                    onClick={() => viewDialog('rules', 'Rules', rules)}
                    style={{ marginRight: 0, textTransform: 'uppercase' }}
                  >
                    View Rules
                  </Button>
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
                  <ViewYamlBtn>
                    <Button
                      onClick={() =>
                        viewDialog('yaml', 'Rules Manifest', manifest)
                      }
                      style={{ marginRight: 0, textTransform: 'uppercase' }}
                    >
                      View Yaml
                    </Button>
                  </ViewYamlBtn>
                ),
              },
            ]}
          />
        </TableWrapper>
      ),
    },
    {
      tabName: 'ROLE BINDINGS',
      data: (
        <TableWrapper id="role-bindings-list">
          <DataTable
            key={listRoleBindings?.objects?.length}
            rows={listRoleBindings?.objects}
            filters={generateFilters(listRoleBindings?.objects)}
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
                  <ViewYamlBtn>
                    <Button
                      onClick={() =>
                        viewDialog('yaml', 'RoleBinding Manifest', manifest)
                      }
                      style={{ marginRight: 0, textTransform: 'uppercase' }}
                    >
                      View Yaml
                    </Button>
                  </ViewYamlBtn>
                ),
              },
            ]}
          />
        </TableWrapper>
      ),
    },
    {
      tabName: 'POLICIES',
      data: (
        <TableWrapper id="workspace-policy-list">
          <DataTable
            key={workspacePolicies?.objects?.length}
            rows={workspacePolicies?.objects}
            filters={generateFilters(workspacePolicies?.objects)}
            fields={[
              {
                label: 'Name',
                value: w => (
                  <Link
                    to={formatURL(Routes.PolicyDetails, {
                      clusterName: clusterName,
                      id: w.id,
                    })}
                    className="link"
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
                defaultSort: true,
                sortValue: ({ createdAt }) => {
                  const t = createdAt && new Date(createdAt).getTime();
                  return t * -1;
                },
              },
            ]}
          />
        </TableWrapper>
      ),
    },
  ];

  return (
    <div style={{minHeight: 'calc(100vh - 335px)'}}>
      <Box>
        <WorkspacesTabs
          className="tabs-container"
          indicatorColor="primary"
          value={selectedTab}
          onChange={handleChange}
          selectionFollowsFocus={true}
        >
          {tabsContent.map(({ tabName }, index) => (
            <WorkspaceTab key={index} className="tab-label" label={tabName} />
          ))}
        </WorkspacesTabs>
      </Box>
      {tabsContent.map((tab, index) => (
        <TabPanel value={selectedTab} index={index} key={index}>
          {tab.data}
        </TabPanel>
      ))}
      {isDialogOpen && (
        <WorkspaceModal
          title={dialogTitle}
          contentType={contentType}
          content={dialogContent || []}
          onFinish={() => setIsDialogOpen(false)}
        />
      )}
    </div>
  );
};

export default TabDetails;
