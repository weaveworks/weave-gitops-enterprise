import {
  Button,
  DataTable,
  Icon,
  IconType,
  filterConfig,
  formatURL,
} from '@weaveworks/weave-gitops';
import { useHistory } from 'react-router-dom';
import styled from 'styled-components';
import { Pipeline } from '../../api/pipelines/types.pb';
import { useListPipelines } from '../../contexts/Pipelines';
import { toFilterQueryString } from '../../utils/FilterQueryString';
import { Routes } from '../../utils/nav';
import { Page } from '../Layout/App';
import { NotificationsWrapper } from '../Layout/NotificationsWrapper';
import { ChipWrapper, LinkWrapper } from '../Policies/PolicyStyles';
import { TableWrapper } from '../Shared';

const Pipelines = ({ className }: any) => {
  const { data, isLoading } = useListPipelines();

  const initialFilterState = {
    ...filterConfig(data?.pipelines, 'namespace'),
  };
  const history = useHistory();

  return (
    <Page loading={isLoading} path={[{ label: 'Pipelines' }]}>
      <NotificationsWrapper errors={data?.errors}>
        <Button
          data-testid="create-pipeline"
          startIcon={<Icon type={IconType.AddIcon} size="base" />}
          onClick={() => {
            history.push(Routes.CreatePipeline);
          }}
        >
          CREATE A PIPELINE
        </Button>
        {data?.pipelines && (
          <TableWrapper className={className} id="pipelines-list">
            <DataTable
              filters={initialFilterState}
              rows={data?.pipelines}
              fields={[
                {
                  label: 'Pipeline Name',
                  value: ({ name, namespace, type }: Pipeline) => (
                    <LinkWrapper
                      to={formatURL(Routes.PipelineDetails, {
                        name,
                        namespace,
                        kind: type,
                      })}
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
                },
                {
                  label: 'Application',
                  value: ({ appRef }: Pipeline) => <>{appRef?.name}</>,
                  sortValue: ({ appRef }: Pipeline) => appRef?.name,
                },
                {
                  label: 'Type',
                  value: ({ appRef }: Pipeline) => <>{appRef?.kind}</>,
                  sortValue: ({ appRef }: Pipeline) => appRef?.kind,
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
                  sortValue: (p: Pipeline) => p.name,
                },
              ]}
            />
          </TableWrapper>
        )}
      </NotificationsWrapper>
    </Page>
  );
};

export default styled(Pipelines).attrs({ className: Pipelines.name })``;
