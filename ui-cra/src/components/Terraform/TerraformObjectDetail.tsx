import { Box } from '@material-ui/core';
import {
  Button,
  Flex,
  formatURL,
  InfoList,
  Interval,
  KubeStatusIndicator,
  Metadata,
  RouterTab,
  SubRouterTabs,
} from '@weaveworks/weave-gitops';
import { useState } from 'react';
import { useRouteMatch } from 'react-router-dom';
import styled from 'styled-components';
import { GetTerraformObjectResponse } from '../../api/terraform/terraform.pb';
import { TerraformObject } from '../../api/terraform/types.pb';
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
import { EditButton } from './../Templates/Edit/EditButton';
import TerraformInventoryTable from './TerraformInventoryTable';

type Props = {
  className?: string;
  name: string;
  namespace: string;
  clusterName: string;
};

const getLabels = (obj: TerraformObject | undefined): [string, string][] => {
  const labels = obj?.labels;
  if (!labels) return [];
  return Object.keys(labels).flatMap(key => {
    return [[key, labels[key] as string]];
  });
};

const getMetadata = (obj: TerraformObject | undefined): [string, string][] => {
  const annotations = obj?.annotations;
  if (!annotations) return [];
  const prefix = 'metadata.weave.works/';
  return Object.keys(annotations).flatMap(key => {
    if (!key.startsWith(prefix)) {
      return [];
    } else {
      return [[key.slice(prefix.length), annotations[key] as string]];
    }
  });
};

function TerraformObjectDetail({ className, ...params }: Props) {
  const { path } = useRouteMatch();
  const [syncing, setSyncing] = useState(false);
  const [suspending, setSuspending] = useState(false);
  const { data, isLoading } = useGetTerraformObjectDetail(params);
  const sync = useSyncTerraformObject(params);
  const toggleSuspend = useToggleSuspendTerraformObject(params);
  const { setNotifications } = useNotifications();

  const { object, yaml, type } = (data || {}) as GetTerraformObjectResponse;

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
            kind: type,
          }),
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
              <Box paddingLeft={1}>
                <EditButton
                  resource={data || ({} as GetTerraformObjectResponse)}
                />
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
                </Box>
                <Box style={{ width: '100%' }}>
                  <TableWrapper>
                    <TerraformInventoryTable rows={object?.inventory || []} />
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
