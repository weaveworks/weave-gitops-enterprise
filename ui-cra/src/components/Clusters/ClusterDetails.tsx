import { ThemeProvider } from '@material-ui/core/styles';
import {
  Button as WeaveButton,
  Icon,
  IconType,
  RouterTab,
  SubRouterTabs,
  theme,
  useListSources,
} from '@weaveworks/weave-gitops';
import { useHistory, useRouteMatch } from 'react-router-dom';
import styled from 'styled-components';
import { localEEMuiTheme } from '../../muiTheme';
import { Routes } from '../../utils/nav';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';

import { CircularProgress, createStyles, makeStyles } from '@material-ui/core';
import { useEffect, useState } from 'react';
import useClusters from '../../hooks/clusters';
import { GitopsClusterEnriched } from '../../types/custom';
import { useIsClusterWithSources } from '../Applications/utils';
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

const useStyles = makeStyles(() =>
  createStyles({
    clusterApplicationBtn: {
      marginBottom: theme.spacing.medium,
    },
    addApplicationBtnLoader: {
      marginLeft: theme.spacing.xl,
    },
  }),
);

const ClusterDetails = ({ clusterName }: Props) => {
  const { path } = useRouteMatch();
  const history = useHistory();
  const classes = useStyles();
  const { isLoading, getCluster, getDashboardAnnotations, getKubeconfig } =
    useClusters();
  const [currentCluster, setCurrentCluster] =
    useState<GitopsClusterEnriched | null>(null);
  const isClusterWithSources = useIsClusterWithSources(clusterName);
  const { isLoading: loading } = useListSources('', '', { retry: false });

  useEffect(
    () => setCurrentCluster(getCluster(clusterName)),
    [clusterName, getCluster],
  );
  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate
        documentTitle="Cluster Page"
        path={[
          { label: 'Clusters', url: Routes.Clusters },
          { label: clusterName },
        ]}
      >
        <ContentWrapper loading={isLoading}>
          {currentCluster && (
            <>
              <ActionsWrapper>
                <WeaveButton
                  id="cluster-application"
                  className={classes.clusterApplicationBtn}
                  startIcon={<Icon type={IconType.FilterIcon} size="base" />}
                  onClick={() => {
                    const filtersValues = encodeURIComponent(
                      `clusterName: ${currentCluster?.namespace}/${currentCluster?.name}`,
                    );
                    history.push(`/applications?filters=${filtersValues}`);
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
            </>
          )}
        </ContentWrapper>
      </PageTemplate>
    </ThemeProvider>
  );
};

export default ClusterDetails;
