import React, { useEffect, useState } from 'react';
import styled from 'styled-components';
import { ThemeProvider } from '@material-ui/core/styles';
import useClusters from '../../hooks/clusters';
import { PageTemplate } from '../Layout/PageTemplate';
import { useHistory, useRouteMatch } from 'react-router-dom';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { localEEMuiTheme } from '../../muiTheme';
import { CAPIClusterStatus } from './CAPIClusterStatus';
import { GitopsClusterEnriched } from '../../types/custom';
import {
  theme,
  InfoList,
  RouterTab,
  SubRouterTabs,
  Icon,
  IconType,
  Button as WeaveButton,
  KubeStatusIndicator,
  useListSources,
} from '@weaveworks/weave-gitops';
import { InfoField } from '@weaveworks/weave-gitops/ui/components/InfoList';
import {
  Box,
  Button,
  makeStyles,
  Typography,
  createStyles,
  CircularProgress,
} from '@material-ui/core';
import { DashboardsList } from './DashboardsList';
import Chip from '@material-ui/core/Chip';
import Divider from '@material-ui/core/Divider';
import { useIsClusterWithSources } from '../Applications/utils';
import { Tooltip } from '../Shared';
import { Routes } from '../../utils/nav';
import { toFilterQueryString } from '../../utils/FilterQueryString';

interface Size {
  size?: 'small';
}

type Props = {
  className?: string;
  name: string;
  namespace: string;
  clusterName: string;
};

const ClusterDashbordWrapper = styled.div`
  .kubeconfig-download {
    padding: 0;
    font-weight: bold;
    color: ${theme.colors.primary};
  }
`;

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

const ClusterDashboard = ({ clusterName }: Props) => {
  const { getCluster, getDashboardAnnotations, getKubeconfig } = useClusters();
  const [currentCluster, setCurrentCluster] =
    useState<GitopsClusterEnriched | null>(null);
  const { path } = useRouteMatch();
  const labels = currentCluster?.labels || {};
  const infrastructureRef = currentCluster?.capiCluster?.infrastructureRef;
  const dashboardAnnotations = getDashboardAnnotations(
    currentCluster as GitopsClusterEnriched,
  );
  const history = useHistory();
  const [disabled, setDisabled] = useState<boolean>(false);
  const classes = useStyles();
  const isClusterWithSources = useIsClusterWithSources(clusterName);
  const { isLoading } = useListSources();

  const handleClick = () => {
    setDisabled(true);
    getKubeconfig(
      clusterName,
      currentCluster?.namespace || '',
      `${clusterName}.kubeconfig`,
    ).finally(() => {
      setDisabled(false);
    });
  };

  const info = [
    [
      'kubeconfig',
      <Button
        className="kubeconfig-download"
        onClick={handleClick}
        disabled={disabled}
      >
        Kubeconfig
      </Button>,
    ],
    ['Namespace', currentCluster?.namespace],
  ];

  const infrastructureRefInfo: InfoField[] = infrastructureRef
    ? [
        ['Kind', infrastructureRef.kind],
        ['APIVersion', infrastructureRef.apiVersion],
      ]
    : [];

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
        <ContentWrapper>
          <ActionsWrapper>
            <WeaveButton
              id="cluster-application"
              className={classes.clusterApplicationBtn}
              startIcon={<Icon type={IconType.FilterIcon} size="base" />}
              onClick={() => {
                const filtersValues = toFilterQueryString([
                  {
                    key: 'clusterName',
                    value: `${currentCluster?.namespace}/${currentCluster?.name}`,
                  },
                ]);
                history.push(`/applications?filters=${filtersValues}`);
              }}
            >
              GO TO APPLICATIONS
            </WeaveButton>

            {isLoading ? (
              <CircularProgress
                size={30}
                className={classes.addApplicationBtnLoader}
              />
            ) : (
              <Tooltip
                title="No sources available for this cluster"
                placement="top"
                disabled={isClusterWithSources === true}
              >
                <div>
                  <WeaveButton
                    id="cluster-add-application"
                    className={classes.clusterApplicationBtn}
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
              <ClusterDashbordWrapper>
                {currentCluster?.conditions &&
                currentCluster?.conditions[0]?.message ? (
                  <div style={{ paddingBottom: theme.spacing.small }}>
                    <KubeStatusIndicator
                      conditions={currentCluster.conditions}
                    />
                  </div>
                ) : null}

                <Box margin={2}>
                  <InfoList items={info as [string, any][]} />
                </Box>
                <Divider variant="middle" />
                {Object.keys(dashboardAnnotations).length > 0 ? (
                  <>
                    <Box margin={2}>
                      <Typography variant="h6" gutterBottom component="div">
                        Dashboards
                      </Typography>
                      <DashboardsList
                        cluster={currentCluster as GitopsClusterEnriched}
                      />
                    </Box>
                    <Divider variant="middle" />
                  </>
                ) : null}
                {Object.keys(labels).length > 0 ? (
                  <>
                    <Box margin={2}>
                      <Typography variant="h6" gutterBottom component="div">
                        Labels
                      </Typography>
                      {Object.entries(labels).map(([key, value]) => (
                        <Chip
                          style={{ marginRight: theme.spacing.small }}
                          key={key}
                          label={`${key}: ${value}`}
                        />
                      ))}
                    </Box>
                    <Divider variant="middle" />
                  </>
                ) : null}
                <Box margin={2}>
                  <CAPIClusterStatus
                    clusterName={clusterName}
                    status={currentCluster?.capiCluster?.status}
                  />
                </Box>
                {infrastructureRef ? (
                  <>
                    <Divider variant="middle" />
                    <Box margin={2}>
                      <Typography variant="h6" gutterBottom component="div">
                        Infrastructure
                      </Typography>
                      <InfoList items={infrastructureRefInfo} />
                    </Box>
                  </>
                ) : null}
              </ClusterDashbordWrapper>
            </RouterTab>
          </SubRouterTabs>
        </ContentWrapper>
      </PageTemplate>
    </ThemeProvider>
  );
};

export default ClusterDashboard;
