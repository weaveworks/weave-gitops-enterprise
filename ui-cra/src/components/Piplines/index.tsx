import { ThemeProvider } from '@material-ui/core/styles';
import { useListPiplines } from '../../contexts/Pipelines';
import { localEEMuiTheme } from '../../muiTheme';
import { useApplicationsCount } from '../Applications/utils';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';

const Piplines = () => {
  const applicationsCount = useApplicationsCount();

  const { error, data, isLoading } = useListPiplines();
  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate documentTitle="WeGo · Piplines">
        <SectionHeader
          className="count-header"
          path={[
            {
              label: 'Applications',
              url: '/applications',
              count: applicationsCount,
            },
            { label: 'Piplines', count: data?.pipelines?.length },
          ]}
        />
        <ContentWrapper loading={isLoading} errorMessage={error?.message}>
          {data?.pipelines && <p>It works</p>}
        </ContentWrapper>
      </PageTemplate>
    </ThemeProvider>
  );
};

export default Piplines;
