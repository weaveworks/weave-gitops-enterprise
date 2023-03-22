import {
  DataTable,
  Flex,
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
};

function initialTerms(search: string) {
  const parsed: { q?: string } = qs.parse(search);

  return parsed.q ? parsed.q.split(',') : [];
}

function Explorer({ className }: Props) {
  const history = useHistory();
  const [queryState, setQueryState] = React.useState<QueryState>({
    query: '',
    pinnedTerms: initialTerms(history.location.search),
    filters: [
      { label: 'Kustomizations', value: 'kind:Kustomization' },
      { label: 'Helm Releases', value: 'kind:HelmRelease' },
    ],
  });
  const { data, error, isFetching } = useQueryService(
    queryState.pinnedTerms.join(','),
  );

  React.useEffect(() => {
    if (queryState.pinnedTerms.length === 0) {
      history.replace(history.location.pathname);
      return;
    }
    const q = qs.stringify({ q: queryState.pinnedTerms.join(',') });

    history.replace(`?${q}`);
  }, [history, queryState.pinnedTerms]);

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
                    pinnedTerms={queryState.pinnedTerms}
                    onChange={(query, pinnedTerms) => {
                      setQueryState({ ...queryState, query, pinnedTerms });
                    }}
                    onPin={pinnedTerms => {
                      setQueryState({ ...queryState, pinnedTerms });
                    }}
                    onFilterSelect={val => {
                      setQueryState({
                        ...queryState,
                        pinnedTerms: [val],
                      });
                    }}
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
