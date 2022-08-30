import { ThemeProvider } from '@material-ui/core/styles';
import { FilterableTable, filterConfig } from '@weaveworks/weave-gitops';
import { Pipeline } from '../../api/pipelines/types.pb';
import { useListPiplines } from '../../contexts/Pipelines';
import { localEEMuiTheme } from '../../muiTheme';
import { useApplicationsCount } from '../Applications/utils';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { LinkWrapper } from '../Policies/PolicyStyles';
import { TableWrapper } from '../Shared';

const Piplines = () => {
  const applicationsCount = useApplicationsCount();
  const { error, data, isLoading } = useListPiplines();

  const initialFilterState = {
    ...filterConfig(data?.pipelines, 'namespace'),
  };

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
          {data?.pipelines && (
            <TableWrapper id="piplines-list">
              <FilterableTable
                filters={initialFilterState}
                rows={data?.pipelines}
                fields={[
                  {
                    label: 'Name',
                    value: (p: Pipeline) => (
                      <LinkWrapper
                        to={`/applications/piplines/details?name=${p.name}`}
                      >
                        {p.name}
                      </LinkWrapper>
                    ),
                  },

                  {
                    label: 'Namespace',
                    value: 'namespace',
                  },
                  {
                    label: 'Application Name',
                    value: ({ appRef }: Pipeline) => <>{appRef?.name}</>,
                  },
                  {
                    label: 'Application Kind',
                    value: ({ appRef }: Pipeline) => <>{appRef?.kind}</>,
                  },
                ]}
              />
            </TableWrapper>
          )}
        </ContentWrapper>
      </PageTemplate>
    </ThemeProvider>
  );
};

export default Piplines;
