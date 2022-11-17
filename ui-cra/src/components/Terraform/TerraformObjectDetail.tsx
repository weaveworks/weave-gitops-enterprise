import { Box } from '@material-ui/core';
import {
  Button,
  DataTable,
  Flex,
  formatURL,
  InfoList,
  Interval,
  KubeStatusIndicator,
  RouterTab,
  SubRouterTabs,
} from '@weaveworks/weave-gitops';
import { useState } from 'react';
import { useRouteMatch } from 'react-router-dom';
import styled from 'styled-components';
import { GetTerraformObjectResponse } from '../../api/terraform/terraform.pb';
import { ResourceRef } from '../../api/terraform/types.pb';
import {
  useGetTerraformObjectDetail,
  useSyncTerraformObject,
  useToggleSuspendTerraformObject,
} from '../../contexts/Terraform';
import { Routes } from '../../utils/nav';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';
import ListEvents from '../ProgressiveDelivery/CanaryDetails/Events/ListEvents';
import { TableWrapper } from '../Shared';
import YamlView from '../YamlView';
import useNotifications from './../../contexts/Notifications';

type Props = {
  className?: string;
  name: string;
  namespace: string;
  clusterName: string;
};

function TerraformObjectDetail({ className, ...params }: Props) {
  const { path } = useRouteMatch();
  const [syncing, setSyncing] = useState(false);
  const [suspending, setSuspending] = useState(false);
  const { data, isLoading } = useGetTerraformObjectDetail(params);
  const sync = useSyncTerraformObject(params);
  const toggleSuspend = useToggleSuspendTerraformObject(params);
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
          url: formatURL(Routes.TerraformDetail, {
            name: object?.name,
            namespace: object?.namespace,
            clusterName: object?.clusterName,
          }),
        },
      ]}
    >
      <ContentWrapper loading={isLoading}>
        <div className={className}>
          <Box paddingBottom={3}>
            <KubeStatusIndicator conditions={object?.conditions || []} />
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
            </Flex>
          </Box>

          <SubRouterTabs rootPath={`${path}/details`}>
            <RouterTab name="Details" path={`${path}/details`}>
              <>
                <Box marginBottom={2}>
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
                      ['Drift Detection Result', object?.driftDetectionResult],
                      ['Suspended', object?.suspended ? 'True' : 'False'],
                    ]}
                  />
                </Box>
                <Box style={{ width: '100%' }}>
                  <TableWrapper>
                    <DataTable
                      fields={[
                        {
                          value: (r: ResourceRef) => r.name as string,
                          label: 'Name',
                        },
                      ]}
                      rows={object?.inventory || []}
                    />
                  </TableWrapper>
                </Box>
              </>
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
            <RouterTab name="Yaml" path={`${path}/yaml`}>
              <>
                <YamlView
                  kind="Terraform"
                  object={{
                    name: object?.name,
                    namespace: object?.namespace,
                  }}
                  yaml={yaml as string}
                />
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
  #events-list {
    width: 100%;
    margin-top: 0;
  }
`;
