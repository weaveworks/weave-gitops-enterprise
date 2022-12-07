import { createStyles, Grid, makeStyles } from '@material-ui/core';
import {
  Flex,
  formatURL,
  Link,
  RouterTab,
  SubRouterTabs,
  theme,
} from '@weaveworks/weave-gitops';
import _ from 'lodash';
import useConfig from '../../../hooks/config';
import styled from 'styled-components';
import {
  Pipeline,
  PipelineTargetStatus,
} from '../../../api/pipelines/types.pb';
import { useGetPipeline } from '../../../contexts/Pipelines';
import { Routes } from '../../../utils/nav';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';
import YamlView from '../../YamlView';
import WorkloadStatus from './WorkloadStatus';
import { EditButton } from './../../../components/Templates/Edit/EditButton';
import { ListError } from '@weaveworks/progressive-delivery/api/prog/types.pb';

const { medium, xs, xxs, large } = theme.spacing;
const { small } = theme.fontSizes;
const { white, neutral10, neutral20, neutral30, black } = theme.colors;

const useStyles = makeStyles(() =>
  createStyles({
    gridContainer: {
      backgroundColor: neutral10,
      margin: `0 ${small}`,
      padding: medium,
      borderRadius: '10px',
    },
    gridWrapper: {
      maxWidth: 'calc(100vw - 300px)',
      display: 'flex',
      flexWrap: 'nowrap',
      overflow: 'auto',
      paddingBottom: '8px',
      margin: `${medium} 0 0 0`,
    },
    title: {
      fontSize: `calc(${small} + ${small})`,
      fontWeight: 600,
      textTransform: 'capitalize',
    },
    subtitle: {
      fontSize: small,
      fontWeight: 400,
      marginTop: xxs,
    },
    mbSmall: {
      marginBottom: small,
    },
    subtitleColor: {
      color: neutral30,
    },
    editButton: {
      paddingBottom: theme.spacing.small,
    },
  }),
);
const TargetWrapper = styled.div`
  font-size: ${theme.fontSizes.large};
  margin-bottom: ${small};
  text-overflow: ellipsis;
  white-space: nowrap;
  overflow: hidden;
  width: calc(250px - ${large});
`;
const CardContainer = styled.div`
  background: ${white};
  padding: ${small};
  margin-bottom: ${xs};
  box-shadow: 0px 0px 1px rgba(26, 32, 36, 0.32);
  border-radius: 10px;
  font-weight: 600;
`;
const Title = styled.div`
  font-size: ${theme.fontSizes.medium};
  color: ${black};
  font-weight: 400;
`;
const ClusterName = styled.div`
  margin-bottom: ${small};
  line-height: 24px;
`;
const TargetNamespace = styled.div`
  font-size: ${theme.fontSizes.medium};
`;
const WorkloadWrapper = styled.div`
  position: relative;
  .version {
    margin-left: ${xxs};
  }
`;
const LastAppliedVersion = styled.span`
  color: ${neutral30};
  font-size: ${theme.fontSizes.medium};
  border: 1px solid ${neutral20};
  padding: 14px 6px;
  border-radius: 50%;
`;
const getTargetsCount = (targetsStatuses: PipelineTargetStatus[]) => {
  return targetsStatuses?.reduce((prev, next) => {
    return prev + (next.workloads?.length || 0);
  }, 0);
};

const mappedErrors = (
  errors: Array<string>,
  namespace: string,
): Array<ListError> => {
  return errors.map(err => ({
    message: err,
    namespace,
  }));
};
interface Props {
  name: string;
  namespace: string;
}

const PipelineDetails = ({ name, namespace }: Props) => {
  const { isLoading, data } = useGetPipeline({
    name,
    namespace,
  });

  const configResponse = useConfig();
  const environments = data?.pipeline?.environments || [];
  const targetsStatuses = data?.pipeline?.status?.environments || {};
  const classes = useStyles();
  const path = `/applications/pipelines/details`;

  return (
    <PageTemplate
      documentTitle="Pipeline Details"
      path={[
        {
          label: 'Applications',
          url: Routes.Applications,
        },
        {
          label: 'Pipelines',
          url: Routes.Pipelines,
        },
        {
          label: name,
        },
      ]}
    >
      <ContentWrapper
        loading={isLoading}
        errors={mappedErrors(data?.errors || [], namespace)}
      >
        <EditButton
          className={classes.editButton}
          resource={data?.pipeline || ({} as Pipeline)}
        />
        <SubRouterTabs rootPath={`${path}/status`}>
          <RouterTab name="Status" path={`${path}/status`}>
            <Grid className={classes.gridWrapper} container spacing={4}>
              {environments.map((env, index) => {
                const status = targetsStatuses[env.name!].targetsStatuses || [];
                return (
                  <Grid
                    item
                    xs
                    key={index}
                    className={classes.gridContainer}
                    id={env.name}
                  >
                    <div className={classes.mbSmall}>
                      <div className={classes.title}>{env.name}</div>
                      <div className={classes.subtitle}>
                        {getTargetsCount(status || [])} Targets
                      </div>
                    </div>
                    {status.map(target => {
                      const clusterName = target?.clusterRef?.name
                        ? `${target?.clusterRef?.namespace}/${target?.clusterRef?.name}`
                        : configResponse?.data?.managementClusterName ||
                          'undefined';

                      return target?.workloads?.map((workload, wrkIndex) => (
                        <CardContainer key={wrkIndex} role="targeting">
                          <TargetWrapper className="workloadTarget">
                            <Title>Cluster</Title>
                            <ClusterName className="cluster-name">
                              {target?.clusterRef?.name || clusterName}
                            </ClusterName>

                            <Title>Namespace</Title>
                            <TargetNamespace className="workload-namespace">
                              {target?.namespace}
                            </TargetNamespace>
                          </TargetWrapper>
                          <WorkloadWrapper>
                            <div>
                              <Link
                                to={formatURL(
                                  '/helm_release/details',
                                  _.omitBy(
                                    {
                                      name: workload?.name,
                                      namespace: target?.namespace,
                                      clusterName,
                                    },
                                    _.isNull,
                                  ),
                                )}
                              >
                                {workload && (
                                  <WorkloadStatus workload={workload} />
                                )}
                              </Link>
                            </div>
                            <Flex wide between>
                              <div
                                style={{ alignSelf: 'flex-end' }}
                                className={`${classes.subtitle} ${classes.subtitleColor}`}
                              >
                                <span>Specification:</span>
                                <span className={`version`}>
                                  {`v${workload?.version}`}
                                </span>
                              </div>
                              {workload?.lastAppliedRevision && (
                                <LastAppliedVersion className="last-applied-version">{`v${workload?.lastAppliedRevision}`}</LastAppliedVersion>
                              )}
                            </Flex>
                          </WorkloadWrapper>
                        </CardContainer>
                      ));
                    })}
                  </Grid>
                );
              })}
            </Grid>
          </RouterTab>
          <RouterTab name="Yaml" path={`${path}/yaml`}>
            <YamlView
              kind="Pipeline"
              yaml={data?.pipeline?.yaml || ''}
              object={data?.pipeline || {}}
            />
          </RouterTab>
        </SubRouterTabs>
      </ContentWrapper>
    </PageTemplate>
  );
};

export default PipelineDetails;
