import { DataTable, Flex } from '@weaveworks/weave-gitops';
import _ from 'lodash';
import qs from 'query-string';
import * as React from 'react';
import { useHistory } from 'react-router-dom';
import styled from 'styled-components';
import { AccessRule } from '../../api/query/query.pb';
import { useListAccessRules, useQueryService } from '../../hooks/query';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';
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
      { label: 'Ready', value: 'status:ready' },
      { label: 'Not Ready', value: 'status:unready' },
    ],
  });
  const { data, error, isFetching } = useQueryService(
    queryState.pinnedTerms.join(','),
  );
  console.log(data);
  const { data: rules } = useListAccessRules();

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
        </div>
        <br />
        <DataTable
          fields={[
            { label: 'Cluster', value: 'cluster' },
            {
              label: 'Subjects',
              value: (r: AccessRule) =>
                _.map(r.subjects, 'name').join(', ') || null,
            },
            {
              label: 'Accessible Kinds',
              value: (r: AccessRule) => r?.accessibleKinds?.join(', ') || null,
            },
          ]}
          rows={rules?.rules}
        />
      </ContentWrapper>
    </PageTemplate>
  );
}

export default styled(Explorer).attrs({ className: Explorer.name })``;
