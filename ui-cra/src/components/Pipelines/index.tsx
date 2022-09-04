import { ThemeProvider } from '@material-ui/core/styles';
import { DataTable, filterConfig } from '@weaveworks/weave-gitops';
import { Pipeline } from '../../api/pipelines/types.pb';
import { useListPipelines } from '../../contexts/Pipelines';
import { localEEMuiTheme } from '../../muiTheme';
import { useApplicationsCount } from '../Applications/utils';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { ChipWrapper, LinkWrapper } from '../Policies/PolicyStyles';
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
              <DataTable
                filters={initialFilterState}
                rows={data?.pipelines}
                fields={[
                  {
                    label: 'Name',
                    value: ({ name, namespace }: Pipeline) => (
                      <LinkWrapper
                        to={`/applications/pipelines/details?namespace=${namespace}&name=${name}`}
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
                    label: 'Kind',
                    value: ({ appRef }: Pipeline) => <>{appRef?.kind}</>,
                    sortValue: ({ appRef }: Pipeline) => appRef?.name,
                  },
                  {
                    label: 'Environments',
                    value: ({ environments }: Pipeline) => (
                      <>
                        {environments?.map(env => (
                          <ChipWrapper key={env.name}>
                            {env.name}
                          </ChipWrapper>
                        ))}
                      </>
                    ),
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
