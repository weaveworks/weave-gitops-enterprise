import { Box } from '@material-ui/core';
import {
  AppContext,
  Button,
  createYamlCommand,
  filterByStatusCallback,
  filterConfig,
  Flex,
  FluxObjectsTable,
  Graph,
  InfoList,
  KubeStatusIndicator,
  Metadata,
  PageStatus,
  ReconciledObjectsAutomation,
  RequestStateHandler,
  RouterTab,
  SubRouterTabs,
  SyncControls,
  YamlView,
} from '@weaveworks/weave-gitops';
import { GroupVersionKind } from '@weaveworks/weave-gitops/ui/lib/api/core/types.pb';
import * as React from 'react';
import { useRouteMatch } from 'react-router-dom';
import styled from 'styled-components';
// Importing this solves a problem with the YAML library not being found.
// @ts-ignore
import * as YAML from 'yaml/browser/dist/index.js';
import { Condition, ObjectRef } from '../../api/gitopssets/types.pb';
import useNotifications from '../../contexts/Notifications';
import {
  useGetGitOpsSet,
  useGetReconciledTree,
  useSyncGitOpsSet,
  useToggleSuspendGitOpsSet,
} from '../../hooks/gitopssets';
import { RequestError } from '../../types/custom';
import { getLabels, getMetadata } from '../../utils/formatters';
import { Routes } from '../../utils/nav';
import { Page } from '../Layout/App';
import { NotificationsWrapper } from '../Layout/NotificationsWrapper';
import ListEvents from '../ListEvents';
import { TableWrapper } from '../Shared';
import { EditButton } from '../Templates/Edit/EditButton';
import { getInventory } from '.';

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

    const suspend = !gs?.suspended;

    return toggleSuspend(suspend)
      .then(() => {
        setNotifications([
          {
            message: {
              text: `Successfully ${suspend ? 'suspended' : 'resumed'} ${
                gs?.name
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

  const { data: gitOpsSet, isLoading: gitOpsSetLoading } = useGetGitOpsSet({
    name,
    namespace,
    clusterName,
  });

  const gs = gitOpsSet?.gitopsSet;

  const {
    data: objects,
    isLoading,
    error,
  } = useGetReconciledTree(
    gs?.name || '',
    gs?.namespace || '',
    (getInventory(gs) as GroupVersionKind[]) || [],
    gs?.clusterName || '',
  );

  const initialFilterState = {
    ...filterConfig(objects, 'type'),
    ...filterConfig(objects, 'namespace'),
    ...filterConfig(objects, 'status', filterByStatusCallback),
  };

  const { setDetailModal } = React.useContext(AppContext);

  if (!gs) {
    return null;
  }

  const reconciledObjectsAutomation: ReconciledObjectsAutomation = {
    source: gs.objectRef || ({} as ObjectRef),
    name: gs.name || '',
    namespace: gs.namespace || '',
    suspended: gs.suspended || false,
    conditions: gs.conditions || ([] as Condition[]),
    type: gs.type || 'GitOpsSet',
    clusterName: gs.clusterName || '',
  };

  const suspended = gs?.suspended;

  return (
    <Page
      loading={gitOpsSetLoading || isLoading}
      path={[
        {
          label: 'GitOpsSet',
          url: Routes.GitOpsSets,
        },
        {
          label: gs?.name || '',
        },
      ]}
    >
      <NotificationsWrapper>
        <Box paddingBottom={3}>
          <KubeStatusIndicator
            conditions={gs?.conditions || []}
            suspended={gs?.suspended}
          />
        </Box>
        <Box paddingBottom={3}>
          <SyncControls
            hideSyncOptions
            syncLoading={syncing}
            syncDisabled={suspended}
            suspendDisabled={suspending || suspended}
            resumeDisabled={suspending || !suspended}
            customActions={[<EditButton resource={gs} />]}
            onSyncClick={handleSyncClick}
            onSuspendClick={handleSuspendClick}
            onResumeClick={handleSuspendClick}
          />
        </Box>
        <SubRouterTabs rootPath={`${path}/details`}>
          <RouterTab name="Details" path={`${path}/details`}>
            <Box style={{ width: '100%' }}>
              <InfoList
                data-testid="info-list"
                items={[
                  ['Observed generation', gs?.observedGeneration],
                  ['Cluster', gs?.clusterName],
                  ['Suspended', suspended ? 'True' : 'False'],
                ]}
              />
              <Metadata metadata={getMetadata(gs)} labels={getLabels(gs)} />
              <TableWrapper>
                <RequestStateHandler
                  loading={isLoading}
                  error={error as RequestError}
                >
                  <FluxObjectsTable
                    className={className}
                    objects={objects || []}
                    onClick={setDetailModal}
                    initialFilterState={initialFilterState}
                  />
                </RequestStateHandler>
              </TableWrapper>
            </Box>
          </RouterTab>
          <RouterTab name="Events" path={`${path}/events`}>
            <ListEvents
              clusterName={gs?.clusterName}
              involvedObject={{
                kind: 'GitOpsSet',
                name: gs?.name,
                namespace: gs?.namespace,
              }}
            />
          </RouterTab>
          <RouterTab name="Graph" path={`${path}/graph`}>
            <RequestStateHandler
              loading={isLoading}
              error={error as RequestError}
            >
              <Graph
                className={className}
                reconciledObjectsAutomation={reconciledObjectsAutomation}
                objects={objects || []}
              />
            </RequestStateHandler>
          </RouterTab>
          <RouterTab name="Yaml" path={`${path}/yaml`}>
            <YamlView
              yaml={YAML.stringify(JSON.parse(gs?.yaml || ('' as string)))}
              type="GitOpsSet"
              header={createYamlCommand(gs?.type, gs?.name, gs?.namespace)}
            />
          </RouterTab>
        </SubRouterTabs>
      </NotificationsWrapper>
    </Page>
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
  .MuiSlider-vertical {
    min-height: 400px;
  }
`;
