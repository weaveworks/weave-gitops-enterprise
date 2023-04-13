import { Box, IconButton } from '@material-ui/core';
import {
  Flex,
  formatURL,
  Icon,
  IconType,
  Link,
  RouterTab,
  SubRouterTabs,
} from '@weaveworks/weave-gitops';
// @ts-ignore
import { DataTable } from '@weaveworks/weave-gitops';
import _ from 'lodash';
import qs from 'query-string';
import * as React from 'react';
import { useHistory } from 'react-router-dom';
import styled from 'styled-components';
import { Object } from '../../api/query/query.pb';
import { useQueryService } from '../../hooks/query';
import { getKindRoute, Routes } from '../../utils/nav';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';
import AccessRulesDebugger from './AccessRulesDebugger';
import QueryBuilder from './QueryBuilder';
import { Field } from '@weaveworks/weave-gitops/ui/components/DataTable';

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
  orderBy: string;
  orderDescending: boolean;
};

function initialTerms(search: string) {
  const parsed: { q?: string } = qs.parse(search);

  return parsed.q ? parsed.q.split(',') : [];
}

const DEFAULT_LIMIT = 25;

// ?clusterName=management&name=flux-system&namespace=flux-system

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
      {
        label: 'Failed',
        value: 'status:Failed',
      },
    ],
    selectedFilter: '',
    orderBy: 'name',
    orderDescending: false,
  });

  const { data, error, isFetching } = useQueryService(
    queryState.pinnedTerms.join(','),
    queryState.limit,
    queryState.offset,
    `${queryState.orderBy} ${queryState.orderDescending ? 'desc' : 'asc'}`,
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
                    {
                      label: 'Name',
                      value: (o: Object) => {
                        const page = getKindRoute(o?.kind as string);

                        const url = formatURL(page, {
                          name: o.name,
                          namespace: o.namespace,
                          clusterName: o.cluster,
                        });

                        return <Link to={url}>{o.name}</Link>;
                      },
                      sortValue: () => 'name',
                    },
                    { label: 'Kind', value: 'kind' },
                    { label: 'Namespace', value: 'namespace' },
                    { label: 'Cluster', value: 'cluster' },
                    {
                      label: 'Status',
                      sortValue: () => 'status',
                      value: (o: Object) => (
                        <Flex align>
                          <Box marginRight={1}>
                            <Icon
                              size={24}
                              color={
                                o?.status === 'Success'
                                  ? 'successOriginal'
                                  : 'alertOriginal'
                              }
                              type={
                                o?.status === 'Success'
                                  ? IconType.SuccessIcon
                                  : IconType.ErrorIcon
                              }
                            />
                          </Box>

                          {o?.status}
                        </Flex>
                      ),
                    },
                    { label: 'Message', value: 'message' },
                  ]}
                  rows={data?.objects}
                  disableSort
                  onColumnHeaderClick={(field: Field) => {
                    const col = _.isFunction(field.value)
                      ? field.sortValue && field.sortValue(field.value)
                      : field.value;

                    setQueryState({
                      ...queryState,
                      orderBy: col as string,
                      orderDescending:
                        queryState.orderBy === col
                          ? !queryState.orderDescending
                          : false,
                    });
                  }}
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

export default styled(Explorer).attrs({ className: Explorer.name })`
  td:last-child {
    white-space: pre-wrap;
    overflow-wrap: break-word;
    word-wrap: break-word;
  }
`;
