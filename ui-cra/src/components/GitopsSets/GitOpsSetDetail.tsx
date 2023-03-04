import { Box } from '@material-ui/core';
import {
  Button,
  Flex,
  InfoList,
  KubeStatusIndicator,
  Metadata,
  PageStatus,
  ReconciledObjectsAutomation,
  ReconciledObjectsTable,
  ReconciliationGraph,
  RouterTab,
  SubRouterTabs,
  YamlView,
} from '@weaveworks/weave-gitops';
import * as React from 'react';
import styled from 'styled-components';
import { useRouteMatch } from 'react-router-dom';
import { Routes } from '../../utils/nav';
import { PageTemplate } from '../Layout/PageTemplate';
import { ContentWrapper } from '../Layout/ContentWrapper';
import ListEvents from '../ProgressiveDelivery/CanaryDetails/Events/ListEvents';
import { TableWrapper } from '../Shared';
import useNotifications from '../../contexts/Notifications';
import {
  useGetReconciledTree,
  useListGitOpsSets,
  useSyncGitOpsSet,
  useToggleSuspendGitOpsSet,
} from '../../hooks/gitopssets';
import { getLabels, getMetadata } from '../../utils/formatters';
import {
  Condition,
  GitOpsSet,
  GroupVersionKind,
  ObjectRef,
} from '../../api/gitopssets/types.pb';
import { getInventory } from '.';

const YAML = require('yaml');

export interface routeTab {
  name: string;
  path: string;
  visible?: boolean;
  component: (param?: any) => any;
}

type Props = {
  className?: string;
  name: string;
  namespace: string;
  clusterName: string;
};

function GitOpsDetail({ className, name, namespace, clusterName }: Props) {
  const { path } = useRouteMatch();
  const [syncing, setSyncing] = React.useState(false);
  const [suspending, setSuspending] = React.useState(false);
  const { data } = useListGitOpsSets();
  const { setNotifications } = useNotifications();

  const sync = useSyncGitOpsSet({
    name,
    namespace,
    clusterName,
  });

  const toggleSuspend = useToggleSuspendGitOpsSet({
    name,
    namespace,
    clusterName,
  });

  const handleSyncClick = () => {
    setSyncing(true);

    return sync()
      .then(() => {
        setNotifications([
          {
            message: { text: 'Sync successful' },
            severity: 'success',
          },
        ]);
      })
      .catch(err => {
        setNotifications([
          {
            message: { text: err?.message },
            severity: 'error',
          },
        ]);
      })
      .finally(() => setSyncing(false));
  };

  const handleSuspendClick = () => {
    setSuspending(true);

    const suspend = !gitOpsSet?.suspended;

    return toggleSuspend(suspend)
      .then(() => {
        setNotifications([
          {
            message: {
              text: `Successfully ${suspend ? 'suspended' : 'resumed'} ${
                gitOpsSet?.name
              }`,
            },
            severity: 'success',
          },
        ]);
      })
      .catch(err => {
        setNotifications([
          { message: { text: err.message }, severity: 'error' },
        ]);
      })
      .finally(() => setSuspending(false));
  };

  const gitOpsSet =
    data?.objects?.find(
      gs =>
        gs.name === name &&
        gs.namespace === namespace &&
        gs.clusterName === clusterName,
    ) || ({} as GitOpsSet);

  const {
    data: objects,
    error,
    isLoading,
  } = useGetReconciledTree(
    gitOpsSet?.name || '',
    gitOpsSet?.namespace || '',
    'GitOpsSet',
    gitOpsSet && (getInventory(gitOpsSet) as GroupVersionKind[]),
    gitOpsSet?.clusterName,
  );

  if (!gitOpsSet) {
    return null;
  }

  const reconciledObjectsAutomation: ReconciledObjectsAutomation = {
    objects: objects || [],
    error: error || undefined,
    isLoading: isLoading || false,
    source: gitOpsSet.objectRef || ({} as ObjectRef),
    name: gitOpsSet.name || '',
    namespace: gitOpsSet.namespace || '',
    suspended: gitOpsSet.suspended || false,
    conditions: gitOpsSet.conditions || ([] as Condition[]),
    type: gitOpsSet.type || 'GitOpsSet',
    clusterName: gitOpsSet.clusterName || '',
  };

  return (
    <PageTemplate
      documentTitle="GitOpsSets"
      path={[
        {
          label: 'GitOpsSet',
          url: Routes.GitOpsSets,
        },
        {
          label: gitOpsSet?.name || '',
        },
      ]}
    >
      <ContentWrapper>
        <Box paddingBottom={3}>
          <KubeStatusIndicator
            conditions={gitOpsSet?.conditions || []}
            suspended={gitOpsSet?.suspended}
          />
        </Box>
        <Box paddingBottom={3}>
          <Flex>
            <Button
              loading={syncing}
              variant="outlined"
              onClick={handleSyncClick}
              style={{ marginRight: 0, textTransform: 'uppercase' }}
            >
              Sync
            </Button>
            <Box paddingLeft={1}>
              <Button
                loading={suspending}
                variant="outlined"
                onClick={handleSuspendClick}
                style={{ marginRight: 0, textTransform: 'uppercase' }}
              >
                {gitOpsSet?.suspended ? 'Resume' : 'Suspend'}
              </Button>
            </Box>
          </Flex>
        </Box>
        <SubRouterTabs rootPath={`${path}/details`}>
          <RouterTab name="Details" path={`${path}/details`}>
            <Box style={{ width: '100%' }}>
              <InfoList
                data-testid="info-list"
                items={[
                  ['Observed generation', gitOpsSet?.observedGeneration],
                  ['Cluster', gitOpsSet?.clusterName],
                  ['Suspended', gitOpsSet?.suspended ? 'True' : 'False'],
                ]}
              />
              <Metadata
                metadata={getMetadata(gitOpsSet)}
                labels={getLabels(gitOpsSet)}
              />
              <TableWrapper>
                <ReconciledObjectsTable
                  reconciledObjectsAutomation={reconciledObjectsAutomation}
                />
              </TableWrapper>
            </Box>
          </RouterTab>
          <RouterTab name="Events" path={`${path}/events`}>
            <ListEvents
              clusterName={gitOpsSet?.clusterName}
              involvedObject={{
                kind: 'GitOpsSet',
                name: gitOpsSet?.name,
                namespace: gitOpsSet?.namespace,
              }}
            />
          </RouterTab>
          <RouterTab name="Graph" path={`${path}/graph`}>
            <ReconciliationGraph
              reconciledObjectsAutomation={reconciledObjectsAutomation}
            />
          </RouterTab>
          <RouterTab name="Yaml" path={`${path}/yaml`}>
            <YamlView
              yaml={YAML.stringify(JSON.parse(gitOpsSet?.yaml as string))}
              object={{
                kind: gitOpsSet?.type,
                name: gitOpsSet?.name,
                namespace: gitOpsSet?.namespace,
              }}
            />
          </RouterTab>
        </SubRouterTabs>
      </ContentWrapper>
    </PageTemplate>
  );
}

export default styled(GitOpsDetail).attrs({
  className: GitOpsDetail?.name,
})`
  ${PageStatus} {
    padding: ${props => props.theme.spacing.small} 0px;
  }
  ${SubRouterTabs} {
    margin-top: ${props => props.theme.spacing.medium};
  }
`;
