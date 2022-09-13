import { DataTable, filterConfig } from '@weaveworks/weave-gitops';
import styled from 'styled-components';
import { Pipeline } from '../../api/pipelines/types.pb';
import { useListPipelines } from '../../contexts/Pipelines';
import { useApplicationsCount } from '../Applications/utils';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { ChipWrapper, LinkWrapper } from '../Policies/PolicyStyles';
import { TableWrapper } from '../Shared';

const Pipelines = ({ className }: any) => {
  const applicationsCount = useApplicationsCount();
  const { error, data, isLoading } = useListPipelines();

  const initialFilterState = {
    ...filterConfig(data?.pipelines, 'namespace'),
  };

  return (
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
          <TableWrapper className={className} id="pipelines-list">
            <DataTable
              filters={initialFilterState}
              rows={data?.pipelines}
              fields={[
                {
                  label: 'Pipeline Name',
                  value: ({ appRef, name, namespace }: Pipeline) => (
                    <LinkWrapper
                      to={`/applications/pipelines/details?namespace=${namespace}&name=${appRef?.name}&pipelineName=${name}`}
                    >
                      {name}
                    </LinkWrapper>
                  ),
                  sortValue: ({ name }: Pipeline) => name,
                  textSearchable: true,
                },
                {
                  label: 'Pipeline Namespace',
                  value: 'namespace',
                  textSearchable: true,
                },
                {
                  label: 'Type',
                  value: ({ appRef }: Pipeline) => <>{appRef?.kind}</>,
                  sortValue: ({ appRef }: Pipeline) => appRef?.name,
                },
                {
                  label: 'Environments',
                  value: ({ environments }: Pipeline) => (
                    <>
                      {environments?.map(env => (
                        <ChipWrapper key={env.name}>{env.name}</ChipWrapper>
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
  );
};

export default styled(Pipelines).attrs({ className: Pipelines.name })``;
