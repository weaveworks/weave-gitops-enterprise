import { Grid } from '@material-ui/core';
import { ThemeProvider } from '@material-ui/core/styles';
import { useCountPipelines, useGetPipeline } from '../../../contexts/Pipelines';
import { localEEMuiTheme } from '../../../muiTheme';
import { useApplicationsCount } from '../../Applications/utils';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';
import { SectionHeader } from '../../Layout/SectionHeader';
import { ChipWrapper, SpaceBetween } from '../../Policies/PolicyStyles';
import { generateRowHeaders, SectionRowHeader } from '../../RowHeader';

const PipelineDetails = ({
  name,
  namespace,
}: {
  name: string;
  namespace: string;
}) => {
  const applicationsCount = useApplicationsCount();
  const pipelinesCount = useCountPipelines();
  const { isLoading, error, data } = useGetPipeline({ name, namespace });

  const defaultHeaders: Array<SectionRowHeader> = [
    {
      rowkey: 'Namespace',
      value: namespace,
    },
    {
      rowkey: 'Application Name',
      value: data?.pipeline?.appRef?.name,
    },
    {
      rowkey: 'Kind',
      value: data?.pipeline?.appRef?.kind,
    },
    {
      rowkey: 'API Version',
      value: data?.pipeline?.appRef?.apiVersion,
    },
  ];

  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate documentTitle="WeGo · Pipeline Details">
        <SectionHeader
          className="count-header"
          path={[
            {
              label: 'Applications',
              url: '/applications',
              count: applicationsCount,
            },
            {
              label: 'Pipelines',
              url: '/applications/pipelines',
              count: pipelinesCount,
            },
            {
              label: name,
            },
          ]}
        />
        <ContentWrapper loading={isLoading} errorMessage={error?.message}>
          {data?.pipeline && (
            <>
              {generateRowHeaders(defaultHeaders)}

              <h2 style={{
                margin:'24px 0'
              }}>Environments</h2>
              <Grid container spacing={3}>
                {Object.entries(data.pipeline.status?.environments || {}).map(
                  ([key, value], index) => (
                    <Grid
                      item
                      xs
                      key={index}
                      style={{
                        boxShadow:'0px 1px 2px rgba(26, 26, 26, 0.24)',
                        margin: '8px',
                        borderRadius: '4px',
                      }}
                    >
                      <div
                        style={{
                          fontSize: '20px',
                          fontWeight: '600',
                          borderBottom: '1px solid',
                        }}
                      >
                        {key}
                      </div>
                      <div
                        className="card-container"
                        style={{
                          paddingTop: '12px',
                        }}
                      >
                        <div
                          style={{
                            fontWeight: '600',
                          }}
                        >
                          Cluster
                        </div>
                        <div
                          style={{
                            padding: '0 8px',
                          }}
                        >
                          <SpaceBetween>
                            <span>{value.clusterRef?.name}</span>
                            <ChipWrapper>{value.clusterRef?.kind}</ChipWrapper>
                          </SpaceBetween>
                        </div>
                      </div>

                      <div
                        className="card-container"
                        style={{
                          paddingTop: '12px',
                        }}
                      >
                        <div
                          style={{
                            fontWeight: '600',
                          }}
                        >
                          Workloads
                        </div>
                        <div
                          style={{
                            padding: '0 8px',
                          }}
                        >
                          {value.workloads?.map((wrk, indx) => (
                            <div key={indx} style={{ marginBottom: '8px' }}>
                              <SpaceBetween>
                                <span>
                                  {wrk.name}@{wrk.version}
                                </span>
                                <ChipWrapper>{wrk.kind}</ChipWrapper>
                              </SpaceBetween>
                            </div>
                          ))}
                        </div>
                      </div>
                    </Grid>
                  ),
                )}
              </Grid>
            </>
          )}
        </ContentWrapper>
      </PageTemplate>
    </ThemeProvider>
  );
};

export default PipelineDetails;
