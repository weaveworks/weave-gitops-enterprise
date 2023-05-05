import { Flex, RouterTab, SubRouterTabs } from '@weaveworks/weave-gitops';
// @ts-ignore
import _ from 'lodash';
import { useState } from 'react';
import styled from 'styled-components';
import { useListFacets, useQueryService } from '../../hooks/query';
import { Routes } from '../../utils/nav';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';
import { TableWrapper } from '../Shared';
import AccessRulesDebugger from './AccessRulesDebugger';
import ExplorerTable from './ExplorerTable';
import FilterDrawer from './FilterDrawer';
import Filters from './Filters';
import PaginationControls from './PaginationControls';
import QueryInput from './QueryInput';
import {
  columnHeaderHandler,
  filterChangeHandler,
  textInputHandler,
  useQueryState,
} from './hooks';

type Props = {
  className?: string;
};

function Explorer({ className }: Props) {
  const [filterDrawerOpen, setFilterDrawerOpen] = useState(false);
  const { data: facetsRes } = useListFacets();
  const [queryState, setQueryState] = useQueryState({
    enableURLState: false,
  });

  const filters = _.reduce(
    queryState.filters,
    (result, f) => {
      const re = /(.+?):\s(.*)/g;

      const matches = re.exec(f);

      if (matches) {
        const [, key, value] = matches;

        result[key] = result[key] || [];
        result[key].push(`${value}`);
      }

      return result;
    },
    {} as any,
  );

  const { data, error, isLoading } = useQueryService({
    terms: queryState.terms,
    filters: queryState.filters,
    limit: queryState.limit,
    offset: queryState.offset,
    orderBy: queryState.orderBy,
    ascending: queryState.orderAscending,
  });

  return (
    <PageTemplate documentTitle="Explorer" path={[{ label: 'Explorer' }]}>
      <ContentWrapper
        loading={isLoading}
        errors={
          error
            ? // Hack to get the message to format correctly.
              // The ContentWrapper API should be simplified to support things other than ListError.
              [{ clusterName: 'Error', message: error?.message }]
            : undefined
        }
      >
        <div className={className}>
          <SubRouterTabs rootPath={`${Routes.Explorer}/query`}>
            <RouterTab name="Query" path={`${Routes.Explorer}/query`}>
              <>
                <TableWrapper>
                  <Flex wide>
                    <ExplorerTable
                      queryState={queryState}
                      rows={data?.objects || []}
                      onColumnHeaderClick={columnHeaderHandler(
                        queryState,
                        setQueryState,
                      )}
                    />
                    <FilterDrawer
                      onClose={() => setFilterDrawerOpen(false)}
                      open={filterDrawerOpen}
                    >
                      <QueryInput
                        queryState={queryState}
                        onTextInputChange={textInputHandler(
                          queryState,
                          setQueryState,
                        )}
                      />

                      <Filters
                        facets={facetsRes?.facets || []}
                        onFilterSelect={filterChangeHandler(
                          queryState,
                          setQueryState,
                        )}
                        state={filters}
                      />
                    </FilterDrawer>
                  </Flex>
                </TableWrapper>

                <PaginationControls
                  queryState={queryState}
                  setQueryState={setQueryState}
                  count={data?.objects?.length || 0}
                />
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

export default styled(Explorer).attrs({ className: Explorer.name })`
  overflow: auto;
`;
