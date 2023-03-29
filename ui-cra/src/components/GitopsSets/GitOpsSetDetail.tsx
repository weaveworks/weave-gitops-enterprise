import { Box } from '@material-ui/core';
import {
  AppContext,
  Button,
  filterByStatusCallback,
  filterConfig,
  Flex,
  FluxObjectsTable,
  InfoList,
  KubeStatusIndicator,
  Metadata,
  PageStatus,
  ReconciledObjectsAutomation,
  RequestStateHandler,
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
  useGetGitOpsSet,
  useGetReconciledTree,
  useSyncGitOpsSet,
  useToggleSuspendGitOpsSet,
} from '../../hooks/gitopssets';
import { getLabels, getMetadata } from '../../utils/formatters';
import { Condition, ObjectRef } from '../../api/gitopssets/types.pb';
import { getInventory } from '.';
import { RequestError } from '../../types/custom';

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
    'GitOpsSet',
    getInventory(gs) || [],
    gs?.clusterName || '',
  );

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

  const initialFilterState = {
    ...filterConfig(objects, 'type'),
    ...filterConfig(objects, 'namespace'),
    ...filterConfig(objects, 'status', filterByStatusCallback),
  };

  const { setDetailModal } = React.useContext(AppContext);

  return (
    <PageTemplate
      documentTitle="GitOpsSets"
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
      <ContentWrapper loading={gitOpsSetLoading || isLoading}>
        <Box paddingBottom={3}>
          <KubeStatusIndicator
            conditions={gs?.conditions || []}
            suspended={gs?.suspended}
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
                {gs?.suspended ? 'Resume' : 'Suspend'}
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
                  ['Observed generation', gs?.observedGeneration],
                  ['Cluster', gs?.clusterName],
                  ['Suspended', gs?.suspended ? 'True' : 'False'],
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
            <Graph
              className={className}
              reconciledObjectsAutomation={reconciledObjectsAutomation}
              objects={objects || []}
            />
          </RouterTab>
          <RouterTab name="Yaml" path={`${path}/yaml`}>
            <YamlView
              yaml={gs?.yaml && YAML.stringify(JSON.parse(gs?.yaml as string))}
              object={{
                kind: gs?.type,
                name: gs?.name,
                namespace: gs?.namespace,
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
