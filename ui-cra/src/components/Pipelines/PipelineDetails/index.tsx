import { Grid } from '@material-ui/core';
import { ThemeProvider } from '@material-ui/core/styles';
import { useCountPipelines, useGetPipeline } from '../../../contexts/Pipelines';
import { localEEMuiTheme } from '../../../muiTheme';
import { useApplicationsCount } from '../../Applications/utils';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';
import { SectionHeader } from '../../Layout/SectionHeader';
import { ChipWrapper } from '../../Policies/PolicyStyles';
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
      rowkey: 'Name',
      value: name,
    },
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
              <div
                style={{
                  margin: 'calc(24px*2) 0',
                }}
              >
                {/* <h2
                  style={{
                    fontSize: '20px',
                    fontWeight: '700',
                  }}
                >
                  Environments
                </h2> */}
                <Grid container spacing={3}>
                  {Object.entries(data.pipeline.status?.environments || {}).map(
                    ([key, value], index) => (
                      <Grid
                        item
                        xs
                        key={index}
                        style={{
                          boxShadow: '0px 1px 2px rgba(26, 26, 26, 0.24)',
                          margin: '8px',
                          borderRadius: '10px',
                        }}
                      >
                        <div
                          style={{
                            fontSize: '20px',
                            fontWeight: '600',
                            borderBottom: '1px solid #f5f5f5',
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
                              fontSize: '12px',
                              color: '#737373',
                            }}
                          >
                            Target
                          </div>
                          <div
                            style={{
                              marginLeft: '8px',
                            }}
                          >
                            <span>
                              {value.clusterRef?.name}/{value.namespace}
                            </span>
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
                              fontSize: '12px',
                              color: '#737373',
                            }}
                          >
                            Name
                          </div>
                          <div
                            style={{
                              marginLeft: '8px',
                            }}
                          >
                            <span>{value.workloads![0].name}</span>
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
                              fontSize: '12px',
                              color: '#737373',
                            }}
                          >
                            Version
                          </div>
                          <div
                            style={{
                              display: 'inline-flex',
                              marginTop: '4px',
                              marginLeft: '8px',
                            }}
                          >
                            {value.workloads?.map((wrk, indx) => (
                              <div key={indx} style={{ marginBottom: '8px' }}>
                                <ChipWrapper>{wrk.version}</ChipWrapper>
                              </div>
                            ))}
                          </div>
                        </div>
                      </Grid>
                    ),
                  )}
                </Grid>
              </div>
            </>
          )}
        </ContentWrapper>
      </PageTemplate>
    </ThemeProvider>
  );
};

export default PipelineDetails;
