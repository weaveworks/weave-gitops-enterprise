import { Box } from '@material-ui/core';
import {
  Button,
  Flex,
  formatURL,
  InfoList,
  Interval,
  KubeStatusIndicator,
  LinkResolverProvider,
  Metadata,
  RouterTab,
  SubRouterTabs,
  YamlView,
} from '@weaveworks/weave-gitops';
import { useState } from 'react';
import { useLocation, useRouteMatch } from 'react-router-dom';
import styled from 'styled-components';
import {
  GetTerraformObjectResponse,
  GetTerraformObjectPlanResponse,
} from '../../api/terraform/terraform.pb';
import {
  useGetTerraformObjectDetail,
  useGetTerraformObjectPlan,
  useSyncTerraformObject,
  useToggleSuspendTerraformObject,
  useReplanTerraformObject,
} from '../../contexts/Terraform';
import { getLabels, getMetadata } from '../../utils/formatters';
import { Routes } from '../../utils/nav';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';
import ListEvents from '../ProgressiveDelivery/CanaryDetails/Events/ListEvents';
import { TableWrapper } from '../Shared';
import useNotifications from './../../contexts/Notifications';
import { EditButton } from './../Templates/Edit/EditButton';
import TerraformDependenciesView from './TerraformDependencyView';
import TerraformInventoryTable from './TerraformInventoryTable';
import TerraformPlanView from './TerraformPlanView';

type Props = {
  className?: string;
  name: string;
  namespace: string;
  clusterName: string;
};

function TerraformObjectDetail({ className, ...params }: Props) {
  const { path } = useRouteMatch();
  const { pathname } = useLocation();
  const [syncing, setSyncing] = useState(false);
  const [suspending, setSuspending] = useState(false);
  const [replanning, setReplanning] = useState(false);
  const { data, isLoading } = useGetTerraformObjectDetail(params);
  const { data: planData, isLoading: isLoadingPlan } =
    useGetTerraformObjectPlan(params);
  const { plan, enablePlanViewing, error } = (planData ||
    {}) as GetTerraformObjectPlanResponse;
  const sync = useSyncTerraformObject(params);
  const toggleSuspend = useToggleSuspendTerraformObject(params);
  const replan = useReplanTerraformObject(params);
  const { setNotifications } = useNotifications();

  const { object, yaml } = (data || {}) as GetTerraformObjectResponse;

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

    const suspend = !object?.suspended;

    return toggleSuspend(suspend)
      .then(() => {
        setNotifications([
          {
            message: {
              text: `Successfully ${suspend ? 'suspended' : 'resumed'} ${
                object?.name
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

  const handleReplanClick = () => {
    setReplanning(true);

    return replan()
      .then(() => {
        setNotifications([
          {
            message: { text: 'Replan requested' },
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
      .finally(() => setReplanning(false));
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

  const shouldShowReplanButton =
    pathname.endsWith('/plan') && !isLoadingPlan && enablePlanViewing && !error;

  return (
    <PageTemplate
      documentTitle="Terraform"
      path={[
        {
          label: 'Terraform Objects',
          url: Routes.TerraformObjects,
        },
        {
          label: params?.name,
        },
      ]}
    >
      <ContentWrapper loading={isLoading}>
        <div className={className}>
          <Box paddingBottom={3}>
            <KubeStatusIndicator
              conditions={object?.conditions || []}
              suspended={object?.suspended}
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
                  {object?.suspended ? 'Resume' : 'Suspend'}
                </Button>
              </Box>
              {shouldShowReplanButton && (
                <Box paddingLeft={1} paddingRight={1}>
                  <Button
                    data-testid="replan-btn"
                    loading={replanning}
                    variant="outlined"
                    onClick={handleReplanClick}
                    style={{ marginRight: 0, textTransform: 'uppercase' }}
                  >
                    Plan
                  </Button>
                </Box>
              )}
              <Box paddingLeft={1}>
                <EditButton
                  resource={data || ({} as GetTerraformObjectResponse)}
                />
              </Box>
            </Flex>
          </Box>
          <SubRouterTabs rootPath={`${path}/details`}>
            <RouterTab name="Details" path={`${path}/details`}>
              <Box style={{ width: '100%' }}>
                <InfoList
                  data-testid="info-list"
                  items={[
                    ['Source', object?.sourceRef?.name],
                    ['Applied Revision', object?.appliedRevision],
                    ['Cluster', object?.clusterName],
                    ['Path', object?.path],
                    [
                      'Interval',
                      <Interval interval={object?.interval as any} />,
                    ],
                    ['Last Update', object?.lastUpdatedAt],
                    [
                      'Drift Detection Result',
                      object?.driftDetectionResult ? 'True' : 'False',
                    ],
                    ['Suspended', object?.suspended ? 'True' : 'False'],
                  ]}
                />
                <Metadata
                  metadata={getMetadata(object)}
                  labels={getLabels(object)}
                />
                <TableWrapper>
                  <TerraformInventoryTable rows={object?.inventory || []} />
                </TableWrapper>
              </Box>
            </RouterTab>
            <RouterTab name="Events" path={`${path}/events`}>
              <ListEvents
                clusterName={object?.clusterName}
                involvedObject={{
                  kind: 'Terraform',
                  name: object?.name,
                  namespace: object?.namespace,
                }}
              />
            </RouterTab>
            <RouterTab name="Dependencies" path={`${path}/dependencies`}>
              <LinkResolverProvider resolver={resolver}>
                <TerraformDependenciesView object={object || {}} />
              </LinkResolverProvider>
            </RouterTab>
            <RouterTab name="Yaml" path={`${path}/yaml`}>
              <YamlView
                yaml={yaml || ''}
                object={{
                  kind: 'Terraform',
                  name: object?.name,
                  namespace: object?.namespace,
                }}
              />
            </RouterTab>
            <RouterTab name="Plan" path={`${path}/plan`}>
              <>
                {!isLoadingPlan && (
                  <TerraformPlanView plan={plan} error={error} />
                )}
              </>
            </RouterTab>
          </SubRouterTabs>
        </div>
      </ContentWrapper>
    </PageTemplate>
  );
}

export default styled(TerraformObjectDetail).attrs({
  className: TerraformObjectDetail.name,
})`
  ${TableWrapper} {
    margin-top: 0;
  }
  #events-list {
    width: 100%;
    margin-top: 0;
  }
`;
