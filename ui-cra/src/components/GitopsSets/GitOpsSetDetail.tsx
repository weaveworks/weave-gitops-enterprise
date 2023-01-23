import { Dialog } from '@material-ui/core';
import {
  AppContext,
  Button,
  CustomActions,
  DependenciesView,
  DialogYamlView,
  EventsTable,
  Flex,
  InfoField,
  InfoList,
  Kind,
  Metadata,
  PageStatus,
  ReconciledObjectsTable,
  ReconciliationGraph,
  RouterTab,
  Spacer,
  SubRouterTabs,
  SyncButton,
  useSyncFluxObject,
  useToggleSuspend,
  YamlView,
} from '@weaveworks/weave-gitops';
import * as React from 'react';
import styled from 'styled-components';
import { useRouteMatch } from 'react-router-dom';
import { GitOpsSet } from '../../api/gitopssets/types.pb';
import Text from './../../components/Text';
import {
  FluxObject,
  FluxObjectNode,
} from '@weaveworks/weave-gitops/ui/lib/objects';

export interface routeTab {
  name: string;
  path: string;
  visible?: boolean;
  component: (param?: any) => any;
}

type Props = {
  gitOpsSet: GitOpsSet;
  className?: string;
  info: InfoField[];
  customTabs?: Array<routeTab>;
  customActions?: JSX.Element[];
};

function GitOpsDetail({
  gitOpsSet,
  className,
  info,
  customTabs,
  customActions,
}: Props) {
  console.log(gitOpsSet);
  const { path } = useRouteMatch();
  const { setNodeYaml, appState } = React.useContext(AppContext);
  const nodeYaml = appState.nodeYaml;
  const sync = useSyncFluxObject([
    {
      name: gitOpsSet.name,
      namespace: gitOpsSet.namespace,
      clusterName: gitOpsSet.clusterName,
      kind: 'GitOpsSet' as Kind,
    },
  ]);

  const suspend = useToggleSuspend(
    {
      objects: [
        {
          name: gitOpsSet.name,
          namespace: gitOpsSet.namespace,
          clusterName: gitOpsSet.clusterName,
          kind: gitOpsSet.type,
        },
      ],
      suspend: !gitOpsSet.suspended,
    },
    'gitOpsSet',
  );

  // default routes
  const defaultTabs: Array<routeTab> = [
    {
      name: 'Details',
      path: `${path}/details`,
      component: () => {
        return (
          <>
            <InfoList items={info} />
            {/* <Metadata metadata={gitOpsSet.metadata} labels={gitOpsSet.labels} /> */}
            <ReconciledObjectsTable automation={gitOpsSet} />
          </>
        );
      },
      visible: true,
    },
    {
      name: 'Events',
      path: `${path}/events`,
      component: () => {
        return (
          <EventsTable
            namespace={gitOpsSet.namespace}
            involvedObject={{
              kind: gitOpsSet.type,
              name: gitOpsSet.name,
              namespace: gitOpsSet.namespace,
              clusterName: gitOpsSet.clusterName,
            }}
          />
        );
      },
      visible: true,
    },
    {
      name: 'Graph',
      path: `${path}/graph`,
      component: () => {
        return (
          <ReconciliationGraph
            parentObject={gitOpsSet}
            source={gitOpsSet.sourceRef}
          />
        );
      },
      visible: true,
    },
    {
      name: 'Dependencies',
      path: `${path}/dependencies`,
      component: () => <DependenciesView automation={gitOpsSet} />,
      visible: true,
    },
    {
      name: 'Yaml',
      path: `${path}/yaml`,
      component: () => {
        return (
          <YamlView
            yaml={gitOpsSet?.generators?.[0] || ''}
            object={{
              kind: 'GitOpsSet' as Kind,
              name: gitOpsSet.name,
              namespace: gitOpsSet.namespace,
            }}
          />
        );
      },
      visible: true,
    },
  ];

  return (
    <Flex wide tall column className={className}>
      <Text size="large" semiBold titleHeight>
        {gitOpsSet.name}
      </Text>
      <PageStatus
        conditions={gitOpsSet.conditions || []}
        suspended={gitOpsSet.suspended}
      />
      <Flex wide start>
        <SyncButton
          onClick={opts => sync.mutateAsync(opts)}
          loading={sync.isLoading}
          disabled={gitOpsSet.suspended}
        />
        <Spacer padding="xs" />
        <Button
          onClick={() => suspend.mutateAsync()}
          loading={suspend.isLoading}
        >
          {gitOpsSet.suspended ? 'Resume' : 'Suspend'}
        </Button>
        {customActions && <CustomActions actions={customActions} />}
      </Flex>

      <SubRouterTabs rootPath={`${path}/details`}>
        {defaultTabs.map(
          (subRoute, index) =>
            subRoute.visible && (
              <RouterTab name={subRoute.name} path={subRoute.path} key={index}>
                {subRoute.component()}
              </RouterTab>
            ),
        )}
        {customTabs?.map(
          (customTab, index) =>
            customTab.visible && (
              <RouterTab
                name={customTab.name}
                path={customTab.path}
                key={index}
              >
                {customTab.component()}
              </RouterTab>
            ),
        )}
      </SubRouterTabs>
      {nodeYaml && (
        <Dialog
          open={!!nodeYaml}
          onClose={() => setNodeYaml({} as FluxObjectNode | FluxObject)}
          maxWidth="md"
          fullWidth
        >
          <DialogYamlView
            object={{
              name: nodeYaml.name,
              namespace: nodeYaml.namespace,
              kind: nodeYaml.type,
            }}
            yaml={nodeYaml.yaml}
          />
        </Dialog>
      )}
    </Flex>
  );
}

export default styled(GitOpsDetail).attrs({
  className: GitOpsDetail.name,
})`
  ${PageStatus} {
    padding: ${props => props.theme.spacing.small} 0px;
  }
  ${SubRouterTabs} {
    margin-top: ${props => props.theme.spacing.medium};
  }
`;
