import { ThemeProvider } from '@material-ui/core/styles';
import { FilterableTable, filterConfig } from '@weaveworks/weave-gitops';
import { Pipeline } from '../../api/pipelines/types.pb';
import { useListPipelines } from '../../contexts/Pipelines';
import { localEEMuiTheme } from '../../muiTheme';
import { useApplicationsCount } from '../Applications/utils';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { LinkWrapper } from '../Policies/PolicyStyles';
import { TableWrapper } from '../Shared';

const Pipelines = () => {
  const applicationsCount = useApplicationsCount();
  const { error, data, isLoading } = useListPipelines();

  const initialFilterState = {
    ...filterConfig(data?.pipelines, 'namespace'),
  };

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
            { label: 'Pipelines', count: data?.pipelines?.length },
          ]}
        />
        <ContentWrapper loading={isLoading} errorMessage={error?.message}>
          {data?.pipelines && (
            <TableWrapper id="pipelines-list">
              <FilterableTable
                filters={initialFilterState}
                rows={data?.pipelines}
                fields={[
                  {
                    label: 'Name',
                    value: ({ name }: Pipeline) => (
                      <LinkWrapper
                        to={`/applications/pipelines/details?name=${name}`}
                      >
                        {name}
                      </LinkWrapper>
                    ),
                    sortValue: ({ name }: Pipeline) => name,
                    textSearchable: true,
                  },
                  {
                    label: 'Namespace',
                    value: 'namespace',
                    textSearchable: true,
                  },
                  {
                    label: 'Application Name',
                    value: ({ appRef }: Pipeline) => <>{appRef?.name}</>,
                    sortValue: ({ appRef }: Pipeline) => appRef?.name,
                  },
                  {
                    label: 'Application Kind',
                    value: ({ appRef }: Pipeline) => <>{appRef?.kind}</>,
                    sortValue: ({ appRef }: Pipeline) => appRef?.name,
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

export default Pipelines;
