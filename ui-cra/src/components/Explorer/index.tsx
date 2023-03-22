import { Box, IconButton } from '@material-ui/core';
import {
  DataTable,
  Flex,
  Icon,
  IconType,
  RouterTab,
  SubRouterTabs,
} from '@weaveworks/weave-gitops';
import qs from 'query-string';
import * as React from 'react';
import { useHistory } from 'react-router-dom';
import styled from 'styled-components';
import { useQueryService } from '../../hooks/query';
import { Routes } from '../../utils/nav';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';
import AccessRulesDebugger from './AccessRulesDebugger';
import QueryBuilder from './QueryBuilder';

type Props = {
  className?: string;
};

type QueryState = {
  query: string;
  pinnedTerms: string[];
  filters: { label: string; value: any }[];
  limit: number;
  offset: number;
  selectedFilter: string;
};

function initialTerms(search: string) {
  const parsed: { q?: string } = qs.parse(search);

  return parsed.q ? parsed.q.split(',') : [];
}

const DEFAULT_LIMIT = 2;

function Explorer({ className }: Props) {
  const history = useHistory();
  const [queryState, setQueryState] = React.useState<QueryState>({
    query: '',
    pinnedTerms: initialTerms(history.location.search),
    limit: DEFAULT_LIMIT,
    offset: 0,
    filters: [
      { label: 'Kustomizations', value: 'kind:Kustomization' },
      { label: 'Helm Releases', value: 'kind:HelmRelease' },
    ],
    selectedFilter: '',
  });
  const { data, error, isFetching } = useQueryService(
    queryState.pinnedTerms.join(','),
    queryState.limit,
    queryState.offset,
  );

  React.useEffect(() => {
    if (queryState.pinnedTerms.length === 0 && queryState.offset === 0) {
      history.replace(history.location.pathname);
      return;
    }

    let offset: any = queryState.offset;
    let limit: any = queryState.limit;

    if (queryState.offset === 0) {
      offset = null;
      limit = null;
    }

    const q = qs.stringify(
      {
        q: queryState.pinnedTerms.join(','),
        limit,
        offset,
      },
      { skipNull: true },
    );

    history.replace(`?${q}`);
  }, [history, queryState.offset, queryState.limit, queryState.pinnedTerms]);

  const handlePageForward = () => {
    setQueryState({
      ...queryState,
      offset: queryState.offset + queryState.limit,
    });
  };

  const handlePageBack = () => {
    setQueryState({
      ...queryState,
      offset: Math.max(0, queryState.offset - queryState.limit),
    });
  };

  const handleFilterChange = (val: string) => {
    let nextVal = [val];

    if (val === '') {
      nextVal = [];
    }

    setQueryState({
      ...queryState,
      offset: 0,
      pinnedTerms: nextVal,
      selectedFilter: val,
    });
  };

  return (
    <PageTemplate documentTitle="Explorer" path={[{ label: 'Explorer' }]}>
      <ContentWrapper
        errors={error ? [{ message: error?.message }] : undefined}
      >
        <div className={className}>
          <SubRouterTabs rootPath={`${Routes.Explorer}/query`}>
            <RouterTab name="Query" path={`${Routes.Explorer}/query`}>
              <>
                <Flex align>
                  <QueryBuilder
                    busy={isFetching}
                    disabled={false}
                    query={queryState.query}
                    filters={queryState.filters}
                    selectedFilter={queryState.selectedFilter}
                    pinnedTerms={queryState.pinnedTerms}
                    onChange={(query, pinnedTerms) => {
                      setQueryState({ ...queryState, query, pinnedTerms });
                    }}
                    onPin={pinnedTerms => {
                      setQueryState({ ...queryState, pinnedTerms });
                    }}
                    onFilterSelect={handleFilterChange}
                  />
                </Flex>

                <DataTable
                  fields={[
                    { label: 'Name', value: 'name' },
                    { label: 'Kind', value: 'kind' },
                    { label: 'Namespace', value: 'namespace' },
                    { label: 'Cluster', value: 'cluster' },
                  ]}
                  rows={data?.objects}
                />
                <Flex wide center>
                  <Box p={2}>
                    <IconButton
                      disabled={queryState.offset === 0}
                      onClick={handlePageBack}
                    >
                      <Icon size={24} type={IconType.NavigateBeforeIcon} />
                    </IconButton>
                    <IconButton
                      disabled={
                        data?.objects &&
                        data?.objects?.length < queryState.limit
                      }
                      onClick={handlePageForward}
                    >
                      <Icon size={24} type={IconType.NavigateNextIcon} />
                    </IconButton>
                  </Box>
                </Flex>
              </>
            </RouterTab>
            <RouterTab name="Access Rules" path={`${Routes.Explorer}/access`}>
              <AccessRulesDebugger />
            </RouterTab>
          </SubRouterTabs>
        </div>
      </ContentWrapper>
    </PageTemplate>
  );
}

export default styled(Explorer).attrs({ className: Explorer.name })``;
