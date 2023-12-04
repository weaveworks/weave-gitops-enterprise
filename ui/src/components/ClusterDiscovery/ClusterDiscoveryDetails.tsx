import { Box } from '@material-ui/core';
import {
  AppContext,
  Button,
  Flex,
  FluxObjectsTable,
  Graph,
  InfoList,
  Kind,
  Page,
  PageStatus,
  RequestStateHandler,
  RouterTab,
  SubRouterTabs,
  YamlView,
  filterByStatusCallback,
  filterConfig,
  useGetInventory,
  useGetObject,
  useToggleSuspend,
  ReconciledObjectsAutomation,
  useSyncFluxObject,
  KubeStatusIndicator,
  createYamlCommand,
  Metadata,
} from '@weaveworks/weave-gitops';
import React from 'react';
import { useRouteMatch } from 'react-router-dom';
import styled from 'styled-components';
import {
  Condition,
  ObjectRef,
} from '../../cluster-services/cluster_services.pb';
import { RequestError } from '../../types/custom';
import { Routes } from '../../utils/nav';
import { NotificationsWrapper } from '../Layout/NotificationsWrapper';
import ListEvents from '../ListEvents';
import { TableWrapper } from '../Shared';

type Props = {
  name: string;
  namespace: string;
  clusterName: string;
  className?: string;
};

function useGetAutomatedClusterDiscovery(
  name: string,
  namespace: string,
  clusterName: string,
) {
  return useGetObject(
    name,
    namespace,
    'AutomatedClusterDiscovery' as Kind,
    clusterName,
  );
}

function ClusterDiscoveryDetails({
  className,
  name,
  namespace,
  clusterName,
}: Props) {
  const { data: acd, error } = useGetAutomatedClusterDiscovery(
    name,
    namespace,
    clusterName,
  );

  const { path } = useRouteMatch();
  const {
    data: invData,
    error: invError,
    isLoading,
  } = useGetInventory(
    'AutomatedClusterDiscovery',
    name,
    clusterName,
    namespace,
  );

  const suspend = useToggleSuspend(
    {
      objects: [
        {
          name,
          namespace,
          clusterName,
          kind: 'AutomatedClusterDiscovery',
        },
      ],
      suspend: !acd?.suspended,
    },
    'object',
  );

  const sync = useSyncFluxObject([
    {
      name,
      namespace,
      clusterName,
      kind: 'AutomatedClusterDiscovery',
    },
  ]);

  const initialFilterState = {
    ...filterConfig(invData, 'type'),
    ...filterConfig(invData, 'namespace'),
    ...filterConfig(invData, 'status', filterByStatusCallback),
  };
  const { setDetailModal } = React.useContext(AppContext);

  if (!acd) {
    return null;
  }

  const { name: nameRef, namespace: namespaceRef } = acd;
  const objectRef = {
    name: nameRef,
    namespace: namespaceRef,
    kind: 'AutomatedClusterDiscovery',
  };
  const reconciledObjectsAutomation: ReconciledObjectsAutomation = {
    source: objectRef || ({} as ObjectRef),
    name: acd.name || '',
    namespace: acd.namespace || '',
    suspended: acd.suspended || false,
    conditions: acd.conditions || ([] as Condition[]),
    type: acd.type || 'AutomatedClusterDiscovery',
    clusterName: acd.clusterName || '',
  };

  return (
    <Page
      loading={isLoading}
      path={[
        {
          label: 'Cluster Discovery',
          url: Routes.ClusterDiscovery,
        },
        {
          label: acd?.name || '',
        },
      ]}
    >
      <NotificationsWrapper>
        <Box paddingBottom={3}>
          <KubeStatusIndicator
            conditions={acd?.conditions || []}
            suspended={acd?.suspended}
          />
        </Box>
        <Box paddingBottom={3}>
          <Flex>
            <Button
              loading={sync.isLoading}
              variant="outlined"
              onClick={() => sync.mutateAsync({ withSource: false })}
              style={{ marginRight: 0, textTransform: 'uppercase' }}
            >
              Sync
            </Button>
            <Box paddingLeft={1}>
              <Button
                loading={suspend.isLoading}
                variant="outlined"
                onClick={() => suspend.mutateAsync()}
                style={{ marginRight: 0, textTransform: 'uppercase' }}
              >
                {acd?.suspended ? 'Resume' : 'Suspend'}
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
                  ['Cluster', acd?.clusterName],
                  ['Suspended', acd?.suspended ? 'True' : 'False'],
                ]}
              />
              <Metadata metadata={acd.metadata} labels={acd.labels} />
              <TableWrapper>
                <RequestStateHandler
                  loading={isLoading}
                  error={error as RequestError}
                >
                  <FluxObjectsTable
                    className={className}
                    objects={invData?.objects || []}
                    onClick={setDetailModal}
                    initialFilterState={initialFilterState}
                  />
                </RequestStateHandler>
              </TableWrapper>
            </Box>
          </RouterTab>
          <RouterTab name="Events" path={`${path}/events`}>
            <ListEvents
              clusterName={acd?.clusterName}
              involvedObject={{
                kind: 'AutomatedClusterDiscovery',
                name: acd?.name,
                namespace: acd?.namespace,
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
                objects={invData?.objects || []}
              />
            </RequestStateHandler>
          </RouterTab>
          <RouterTab name="Yaml" path={`${path}/yaml`}>
            <YamlView
              yaml={acd?.yaml}
              header={createYamlCommand(acd?.type, acd?.name, acd?.namespace)}
            />
          </RouterTab>
        </SubRouterTabs>
      </NotificationsWrapper>
    </Page>
  );
}

export default styled(ClusterDiscoveryDetails).attrs({
  className: ClusterDiscoveryDetails?.name,
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
