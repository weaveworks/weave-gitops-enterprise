import { Alert } from '@material-ui/lab';
import _ from 'lodash';
import styled from 'styled-components';
import { useListFacets, useQueryService } from '../../hooks/query';
import ExplorerTable from './ExplorerTable';
import PaginationControls from './PaginationControls';
import QueryInput from './QueryInput';
import { columnHeaderHandler, textInputHandler, useQueryState } from './hooks';

type Props = {
  className?: string;
  enableBatchSync?: boolean;
  category: 'automation' | 'source';
};

const categoryKinds = {
  automation: ['Kustomization', 'HelmRelease'],
  source: [
    'GitRepository',
    'HelmRepository',
    'Bucket',
    'HelmChart',
    'OCIRepository',
  ],
};

function ScopedExploreUI({ className, category, enableBatchSync }: Props) {
  const { data: facets } = useListFacets();
  const [queryState, setQueryState] = useQueryState({
    enableURLState: false,
  });

  const { data, error, isLoading } = useQueryService({
    terms: queryState.terms,
    filters: queryState.filters,
    limit: queryState.limit,
    offset: queryState.offset,
    orderBy: queryState.orderBy,
    ascending: queryState.orderAscending,
    category,
  });

  if (isLoading) {
    return null;
  }

  if (error) {
    return <Alert severity="error">Error: {error.message}</Alert>;
  }

  // We only want certain kinds to show up as facets
  const withoutKind = _.filter(facets?.facets, facet => {
    return facet.field?.toLowerCase() !== 'kind';
  });

  const kindFacets = _.map(categoryKinds[category], k => k.toLowerCase());

  withoutKind.push({
    field: 'Kind',
    values: kindFacets,
  });

  return (
    <div className={className}>
      <QueryInput
        queryState={queryState}
        onTextInputChange={textInputHandler(queryState, setQueryState)}
      />
      <ExplorerTable
        queryState={queryState}
        className={className}
        rows={data?.objects || []}
        onColumnHeaderClick={columnHeaderHandler(queryState, setQueryState)}
        enableBatchSync={enableBatchSync}
        sortField={queryState.orderBy}
      />
      <PaginationControls
        queryState={queryState}
        setQueryState={setQueryState}
        count={data?.objects?.length || 0}
      />
    </div>
  );
}

export default styled(ScopedExploreUI).attrs({
  className: ScopedExploreUI.name,
})``;
