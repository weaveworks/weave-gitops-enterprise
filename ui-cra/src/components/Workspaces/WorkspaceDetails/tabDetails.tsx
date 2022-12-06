import React, { useState } from 'react';
import styled from 'styled-components';
import {
  Tab,
  Tabs,
  Box,
} from '@material-ui/core';
import {
  Button,
  DataTable,
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

const WorkspacesTabs = styled(Tabs)`
  min-height: 32px !important;
  .link{
    color: ${({ theme }) => theme.colors.primary},
    fontWeight: 600,
    whiteSpace: 'pre-line',
  }
`;

const WorkspaceTab = styled(Tab)(({ theme }) => ({
  '&.MuiTab-root': {
    fontSize: theme.fontSizes.small,
    fontWeight: 600,
    minHeight: '32px',
    minWidth: '133px',
    opacity: 1,
    paddingLeft: '0 !important',
    paddingRight: '0 !important',
    span: {
      color: theme.colors.neutral30,
    },
  },
  '&.Mui-selected': {
    fontWeight: 700,
    background: `${theme.colors.primary}1A`,
    span: {
      color: theme.colors.primary10,
    },
  },
  '&.Mui-focusVisible': {
    backgroundColor: '#d1eaff',
  },
}));

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
      {value === index && <Box sx={{ p: 3 }}>{children}</Box>}
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
  const [value, setValue] = useState<number>(0);
  const [dialogContent, setDialogContent] = useState<
    string | WorkspaceRoleRule[]
  >();
  const [isDialogOpen, setIsDialogOpen] = useState<boolean>(false);
  const [contentType, setContentType] = useState<string>('');
  const [dialogTitle, setDialogTitle] = useState<string>('');

  const { data: roles} = useGetWorkspaceRoles({
    clusterName,
    workspaceName,
  });

  const { data: listRoleBindings} =
    useGetWorkspaceRoleBinding({
      clusterName,
      workspaceName,
    });

  const { data: serviceAccounts }=
    useGetWorkspaceServiceAccount({
      clusterName,
      workspaceName,
    });
  const { data: workspacePolicies} =
    useGetWorkspacePolicies({
      clusterName,
      workspaceName,
    });

  const handleChange = (event: React.ChangeEvent<{}>, newValue: number) => {
    setValue(newValue);
  };
  const tabsContent: Array<TabContent> = [
    {
      tabName: 'SERVICE ACCOUNTS',
      data: (
        <TableWrapper id="service-accounts-list">
          <DataTable
            key={serviceAccounts?.objects?.length}
            rows={serviceAccounts?.objects}
            fields={[
              {
                label: 'Name',
                value: 'name',
                textSearchable: true,
                sortValue: ({ name }) => name,
                maxWidth: 650,
              },
              {
                label: 'Namespace',
                value: 'namespace',
              },
              {
                label: 'Age',
                value: ({ timestamp }) => moment(timestamp).fromNow(),
              },
              {
                label: '',
                value: ({ manifest }) => (
                  <Button
                    onClick={() => {
                      setContentType('yaml');
                      setDialogTitle('Service Accounts Manifest');
                      setDialogContent(manifest);
                      setIsDialogOpen(true);
                    }}
                    style={{ marginRight: 0, textTransform: 'uppercase' }}
                  >
                    View Yaml
                  </Button>
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
            fields={[
              {
                label: 'Name',
                value: 'name',
                textSearchable: true,
                sortValue: ({ name }) => name,
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
                    onClick={() => {
                      setContentType('rules');
                      setDialogTitle('Rules');
                      setDialogContent(rules);
                      setIsDialogOpen(true);
                    }}
                    style={{ marginRight: 0, textTransform: 'uppercase' }}
                  >
                    View Rules
                  </Button>
                ),
              },
              {
                label: 'Age',
                value: ({ timestamp }) => moment(timestamp).fromNow(),
              },
              {
                label: '',
                value: ({ manifest }) => (
                  <Button
                    onClick={() => {
                      setContentType('yaml');
                      setDialogTitle('Rules Manifest');
                      setDialogContent(manifest);
                      setIsDialogOpen(true);
                    }}
                    style={{ marginRight: 0, textTransform: 'uppercase' }}
                  >
                    View Yaml
                  </Button>
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
            fields={[
              {
                label: 'Name',
                value: 'name',
                textSearchable: true,
                sortValue: ({ name }) => name,
                maxWidth: 650,
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
              },
              {
                label: '',
                value: ({ manifest }) => (
                  <Button
                    onClick={() => {
                      setContentType('yaml');
                      setDialogTitle('RoleBinding Manifest');
                      setDialogContent(manifest);
                      setIsDialogOpen(true);
                    }}
                    style={{ marginRight: 0, textTransform: 'uppercase' }}
                  >
                    View Yaml
                  </Button>
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
        <TableWrapper id="policy-list">
          <DataTable
            key={workspacePolicies?.objects?.length}
            rows={workspacePolicies?.objects}
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
              },
            ]}
          />
        </TableWrapper>
      ),
    },
  ];
  console.log(workspacePolicies);
  const onFinish = () => {
    setIsDialogOpen(false);
  };

  return (
    <>
      <Box>
        <WorkspacesTabs
          className="tabs-container"
          indicatorColor="primary"
          value={value}
          onChange={handleChange}
          aria-label="pr-preview-sections"
          selectionFollowsFocus={true}
        >
          {tabsContent.map(({ tabName, value }, index) => (
            <WorkspaceTab key={index} className="tab-label" label={tabName} />
          ))}
        </WorkspacesTabs>
      </Box>
      {tabsContent.map((tab, index) => (
        <TabPanel value={value} index={index} key={index}>
          {tab.data}
        </TabPanel>
      ))}
      {isDialogOpen && (
        <WorkspaceModal
          title={dialogTitle}
          contentType={contentType}
          content={dialogContent || ''}
          onFinish={onFinish}
        />
      )}
    </>
  );
};

export default TabDetails;
