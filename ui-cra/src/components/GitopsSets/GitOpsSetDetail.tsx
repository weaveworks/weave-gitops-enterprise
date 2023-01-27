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
import { ObjectRef } from '@weaveworks/weave-gitops/ui/lib/api/core/types.pb';
import { Routes } from '../../utils/nav';
import { PageTemplate } from '../Layout/PageTemplate';
import { ContentWrapper } from '../Layout/ContentWrapper';
import ListEvents from '../ProgressiveDelivery/CanaryDetails/Events/ListEvents';

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
  // remove hardcoded object only clusterName is sorted out
  var gitOpsSet = {
    name: 'gitopsset-configmaps',
    namespace: 'default',
    inventory: [
      { id: 'default_dev-info-configmap__ConfigMap', version: 'v1' },
      { id: 'default_production-info-configmap__ConfigMap', version: 'v1' },
      { id: 'default_staging-info-configmap__ConfigMap', version: 'v1' },
    ],
    conditions: [
      {
        type: 'Ready',
        status: 'True',
        reason: 'ReconciliationSucceeded',
        message: '3 resources created',
        timestamp: '2023-01-24 13:27:17 +0000 UTC',
      },
    ],
    generators: [
      '{"elements":[{"env":"dev","team":"dev-team"},{"env":"production","team":"ops-team"},{"env":"staging","team":"ops-team"}]}',
    ],
    clusterName: 'management',
    type: 'GitOpsSet',
    labels: {},
    annotations: {
      'kubectl.kubernetes.io/last-applied-configuration':
        '{"apiVersion":"templates.weave.works/v1alpha1","kind":"GitOpsSet","metadata":{"annotations":{},"name":"gitopsset-configmaps","namespace":"default"},"spec":{"generators":[{"list":{"elements":[{"env":"dev","team":"dev-team"},{"env":"production","team":"ops-team"},{"env":"staging","team":"ops-team"}]}}],"templates":[{"apiVersion":"v1","kind":"ConfigMap","metadata":{"name":"{{ .env }}-info-configmap","namespace":"default"},"spec":{"data":{"description":"This is a configmap for the {{ .env }} environment","env":"{{ .env }}","team":"{{ .team }}"}}}]}}\n',
    },
    sourceRef: {
      apiVersion: 'templates.weave.works/v1alpha1',
      kind: 'GitOpsSet',
      name: 'gitopsset-configmaps',
      namespace: 'default',
    },
    suspend: false,
  } as GitOpsSet;
  console.log(gitOpsSet);
  const { path } = useRouteMatch();
  const { setNodeYaml, appState } = React.useContext(AppContext);
  const nodeYaml = appState.nodeYaml;
  // const sync = useSyncFluxObject([
  //   {
  //     name: gitOpsSet.name,
  //     namespace: gitOpsSet.namespace,
  //     clusterName: gitOpsSet.clusterName,
  //     kind: 'GitOpsSet' as Kind,
  //   },
  // ]);

  // const suspend = useToggleSuspend(
  //   {
  //     objects: [
  //       {
  //         name: gitOpsSet.name,
  //         namespace: gitOpsSet.namespace,
  //         clusterName: gitOpsSet.clusterName,
  //         kind: gitOpsSet.type,
  //       },
  //     ],
  //     suspend: !gitOpsSet.suspended,
  //   },
  //   'gitOpsSet',
  // );

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
            {/* <ReconciledObjectsTable automation={gitOpsSet} /> */}
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
          <ListEvents
            involvedObject={{
              kind: 'GitOpsSet',
              name: gitOpsSet.name,
              namespace: gitOpsSet.namespace,
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
            source={gitOpsSet.sourceRef || ({} as ObjectRef)}
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
        <Flex wide tall column className={className}>
          <Text size="large" semiBold titleHeight>
            {gitOpsSet.name}
          </Text>
          <PageStatus
            conditions={gitOpsSet.conditions || []}
            suspended={gitOpsSet.suspended || false}
          />
          <Flex wide start>
            {/* <SyncButton
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
            </Button> */}
            {customActions && <CustomActions actions={customActions} />}
          </Flex>

          <SubRouterTabs rootPath={`${path}/details`}>
            {defaultTabs.map(
              (subRoute, index) =>
                subRoute.visible && (
                  <RouterTab
                    name={subRoute.name}
                    path={subRoute.path}
                    key={index}
                  >
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
      </ContentWrapper>
    </PageTemplate>
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
