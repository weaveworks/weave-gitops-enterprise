import { ThemeProvider } from '@material-ui/core/styles';
import { useCountPipelines, useGetPipeline } from '../../../contexts/Pipelines';
import { localEEMuiTheme } from '../../../muiTheme';
import { useApplicationsCount } from '../../Applications/utils';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';
import { SectionHeader } from '../../Layout/SectionHeader';
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

  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate documentTitle="WeGo · Pipelines">
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
          ]}
        />
        <ContentWrapper loading={isLoading} errorMessage={error?.message}>
          {data?.pipeline && <>Pipeline Details</>}
        </ContentWrapper>
      </PageTemplate>
    </ThemeProvider>
  );
};

export default PipelineDetails;
