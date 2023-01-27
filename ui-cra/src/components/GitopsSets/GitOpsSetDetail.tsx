import { Box, Dialog } from '@material-ui/core';
import {
  AppContext,
  Button,
  Flex,
  formatURL,
  InfoList,
  KubeStatusIndicator,
  LinkResolverProvider,
  Metadata,
  PageStatus,
  RouterTab,
  SubRouterTabs,
} from '@weaveworks/weave-gitops';
import * as React from 'react';
import styled from 'styled-components';
import { useRouteMatch } from 'react-router-dom';
import { GitOpsSet } from '../../api/gitopssets/types.pb';
import { ObjectRef } from '@weaveworks/weave-gitops/ui/lib/api/core/types.pb';
import { Routes } from '../../utils/nav';
import { PageTemplate } from '../Layout/PageTemplate';
import { ContentWrapper } from '../Layout/ContentWrapper';
import ListEvents from '../ProgressiveDelivery/CanaryDetails/Events/ListEvents';
import CodeView from '../CodeView';
import { TableWrapper } from '../Shared';
import useNotifications from '../../contexts/Notifications';
import {
  useListGitOpsSets,
  useToggleSuspendGitOpsSet,
} from '../../contexts/GitOpsSets';
import { getLabels, getMetadata } from '../../utils/formatters';
import GitOpsSetInventoryTable from './GitOpsSetInventoryTable';

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

  const gitOpsSet = data?.gitopssets?.find(
    gs =>
      gs.name === name &&
      gs.namespace === namespace &&
      gs.clusterName === clusterName,
  );
  // const sync = useSyncGitOpsSet({
  //   name,
  //   namespace,
  //   clusterName,
  // });
  const toggleSuspend = useToggleSuspendGitOpsSet({
    name,
    namespace,
    clusterName,
  });
  const { setNotifications } = useNotifications();

  // const handleSyncClick = () => {
  //   setSyncing(true);

  //   return sync()
  //     .then(() => {
  //       setNotifications([
  //         {
  //           message: { text: 'Sync successful' },
  //           severity: 'success',
  //         },
  //       ]);
  //     })
  //     .catch(err => {
  //       setNotifications([
  //         {
  //           message: { text: err?.message },
  //           severity: 'error',
  //         },
  //       ]);
  //     })
  //     .finally(() => setSyncing(false));
  // };

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

  const resolver = (type: string, params: any) => {
    return (
      formatURL(Routes.TerraformDetail, {
        name: params.name,
        namespace: params.namespace,
        clusterName: params.clusterName,
      }) || ''
    );
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
              // onClick={handleSyncClick}
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
                  ['Source', gitOpsSet?.sourceRef?.name],
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
                <GitOpsSetInventoryTable rows={gitOpsSet?.inventory || []} />
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
          <RouterTab name="Dependencies" path={`${path}/dependencies`}>
            <LinkResolverProvider resolver={resolver}>
              {/* this will need depends on */}
              {/* <TerraformDependenciesView object={gitOpsSet || {}} /> */}
            </LinkResolverProvider>
          </RouterTab>
          <RouterTab name="Yaml" path={`${path}/yaml`}>
            <CodeView
              kind="GitOpsSet"
              object={{
                name: gitOpsSet?.name,
                namespace: gitOpsSet?.namespace,
              }}
              //we should be able to handle an array
              code={gitOpsSet?.generators?.[0] || ''}
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
