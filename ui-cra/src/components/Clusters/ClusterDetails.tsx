import { ThemeProvider } from '@material-ui/core/styles';
import {
  Icon,
  IconType,
  Page,
  RouterTab,
  SubRouterTabs,
  Button as WeaveButton,
  useFeatureFlags,
  useListSources,
} from '@weaveworks/weave-gitops';
import { useHistory, useRouteMatch } from 'react-router-dom';
import styled from 'styled-components';
import { localEEMuiTheme } from '../../muiTheme';
import { Routes } from '../../utils/nav';
import { CircularProgress } from '@material-ui/core';
import { useEffect, useState } from 'react';
import useClusters from '../../hooks/clusters';
import { GitopsClusterEnriched } from '../../types/custom';
import { toFilterQueryString } from '../../utils/FilterQueryString';
import { useIsClusterWithSources } from '../Applications/utils';
import { QueryState } from '../Explorer/hooks';
import { linkToExplorer } from '../Explorer/utils';
import { Tooltip } from '../Shared';
import ClusterDashboard from './ClusterDashboard';
type Props = {
  className?: string;
  name: string;
  namespace: string;
  clusterName: string;
};
interface Size {
  size?: 'small';
}
const ActionsWrapper = styled.div<Size>`
  display: flex;
`;

const ClusterDetails = ({ clusterName }: Props) => {
  const { path } = useRouteMatch();
  const history = useHistory();
  const { isLoading, getCluster, getDashboardAnnotations, getKubeconfig } =
    useClusters();
  const [currentCluster, setCurrentCluster] =
    useState<GitopsClusterEnriched | null>(null);
  const isClusterWithSources = useIsClusterWithSources(clusterName);
  const { isLoading: loading } = useListSources('', '', { retry: false });

  const { isFlagEnabled } = useFeatureFlags();
  const useQueryServiceBackend = isFlagEnabled(
    'WEAVE_GITOPS_FEATURE_QUERY_SERVICE_BACKEND',
  );

  useEffect(
    () => setCurrentCluster(getCluster(clusterName)),
    [clusterName, getCluster],
  );
  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <Page
        loading={isLoading}
        path={[
          { label: 'Clusters', url: Routes.Clusters },
          { label: clusterName },
        ]}
      >
        {currentCluster && (
          <div style={{ overflowX: 'auto' }}>
            <ActionsWrapper style={{ marginBottom: 8 }}>
              <WeaveButton
                id="cluster-application"
                startIcon={<Icon type={IconType.FilterIcon} size="base" />}
                onClick={() => {
                  const clusterName = `${currentCluster?.namespace}/${currentCluster?.name}`;
                  if (useQueryServiceBackend) {
                    const s = linkToExplorer(`/applications`, {
                      filters: [`Cluster:${clusterName}`],
                    } as QueryState);

                    history.push(s);
                  } else {
                    const filtersValues = toFilterQueryString([
                      {
                        key: 'clusterName',
                        value: `${currentCluster?.namespace}/${currentCluster?.name}`,
                      },
                    ]);
                    history.push(`/applications?filters=${filtersValues}`);
                  }
                }}
              >
                GO TO APPLICATIONS
              </WeaveButton>
              {loading ? (
                <CircularProgress size={30} />
              ) : (
                <Tooltip
                  title="No sources available for this cluster"
                  placement="top"
                  disabled={isClusterWithSources === true}
                >
                  <div>
                    <WeaveButton
                      id="cluster-add-application"
                      startIcon={<Icon type={IconType.AddIcon} size="base" />}
                      onClick={() => {
                        const filtersValues = encodeURIComponent(
                          `${currentCluster?.name}`,
                        );
                        history.push(
                          `/applications/create?clusterName=${filtersValues}`,
                        );
                      }}
                      disabled={!isClusterWithSources}
                    >
                      ADD APPLICATION TO THIS CLUSTER
                    </WeaveButton>
                  </div>
                </Tooltip>
              )}
            </ActionsWrapper>
            <SubRouterTabs rootPath={`${path}/details`}>
              <RouterTab name="Details" path={`${path}/details`}>
                <ClusterDashboard
                  currentCluster={currentCluster}
                  getDashboardAnnotations={getDashboardAnnotations}
                  getKubeconfig={getKubeconfig}
                />
              </RouterTab>
            </SubRouterTabs>
          </div>
        )}{' '}
      </Page>
    </ThemeProvider>
  );
};

export default ClusterDetails;
