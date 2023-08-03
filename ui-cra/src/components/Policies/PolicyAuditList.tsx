import { Alert } from '@material-ui/lab';
import {
  Button,
  DataTable,
  Flex,
  Icon,
  IconType,
  Link,
  LoadingPage,
  MessageBox,
  Severity,
  Spacer,
  Text,
  Timestamp,
  V2Routes,
  formatURL,
  useFeatureFlags
} from '@weaveworks/weave-gitops';
import { useHistory } from 'react-router-dom';
import { useListFacets, useQueryService } from '../../hooks/query';
import { URLQueryStateManager } from '../Explorer/QueryStateManager';
import { QueryStateProvider, columnHeaderHandler } from '../Explorer/hooks';

// @ts-ignore
import { IconButton } from '@material-ui/core';
import { useState } from 'react';
import FilterDrawer from '../Explorer/FilterDrawer';
import Filters from '../Explorer/Filters';
import PaginationControls from '../Explorer/PaginationControls';
import QueryInput from '../Explorer/QueryInput';
import QueryStateChips from '../Explorer/QueryStateChips';
import { NotificationsWrapper } from '../Layout/NotificationsWrapper';
import { LinkTag, TableWrapper } from '../Shared';

const PolicyAuditList = () => {
  const [filterDrawerOpen, setFilterDrawerOpen] = useState(false);
  const { isFlagEnabled } = useFeatureFlags();

  const useQueryServiceBackend = isFlagEnabled(
    'WEAVE_GITOPS_FEATURE_QUERY_SERVICE_BACKEND',
  );

  const history = useHistory();
  const manager = new URLQueryStateManager(history);
  const queryState = manager.read();
  const setQueryState = manager.write;

  const { data: facetsRes } = useListFacets();
  const filteredFacets = facetsRes?.facets?.filter(f => f.field !== 'kind');

  const { data, error, isLoading } = useQueryService({
    terms: queryState.terms,
    filters: ['kind:Event', ...queryState.filters],
    limit: queryState.limit,
    offset: queryState.offset,
    orderBy: queryState.orderBy,
    ascending: queryState.orderAscending,
  });

  const rows = data?.objects?.map(obj => {
    const { unstructured, cluster } = obj;
    const details = JSON.parse(unstructured || '');
    const {
      metadata: {
        annotations: { category, policy_name, severity, policy_id },
        creationTimestamp,
      },
      involvedObject: { namespace, name },
      message,
    } = details;
    return {
      message,
      cluster,
      category,
      policy_name,
      severity,
      creationTimestamp,
      policy_id,
      namespace,
      name,
    };
  });

  return useQueryServiceBackend ? (
    <QueryStateProvider manager={manager}>
      {error && <Alert severity="error">{error.message}</Alert>}
      {isLoading ? (
        <LoadingPage />
      ) : (
        <>
          <Flex align wide end>
            <QueryStateChips />
            <IconButton onClick={() => setFilterDrawerOpen(!filterDrawerOpen)}>
              <Icon
                size="normal"
                type={IconType.FilterIcon}
                color="neutral30"
              />
            </IconButton>
          </Flex>
          <Flex wide>
            <TableWrapper id="auditViolations-list">
              <DataTable
                key={rows?.length}
                rows={rows}
                fields={[
                  {
                    label: 'Message',
                    value: ({ message }) => (
                      <span title={message}>{message}</span>
                    ),
                    textSearchable: true,
                    maxWidth: 300,
                  },
                  {
                    label: 'Cluster',
                    value: 'cluster',
                  },
                  {
                    label: 'Application',
                    value: ({ namespace, name }) => `${namespace}/${name}`,
                    sortValue: ({ namespace, name }) => `${namespace}/${name}`,
                    maxWidth: 150,
                  },
                  {
                    label: 'Severity',
                    value: ({ severity }) => (
                      <Severity severity={severity || ''} />
                    ),
                    sortValue: ({ severity }) => severity,
                  },
                  {
                    label: 'Category',
                    value: ({ category }) => (
                      <span title={category}>{category}</span>
                    ),
                    sortValue: ({ category }) => category,
                    maxWidth: 100,
                  },

                  {
                    label: 'Violated Policy',
                    value: ({ policy_name, cluster, policy_id }) => (
                      <Link
                        to={formatURL(V2Routes.PolicyDetailsPage, {
                          clusterName: cluster,
                          id: policy_id,
                          name: policy_name,
                        })}
                        data-policy-name={policy_name}
                      >
                        <Text capitalize semiBold>
                          {policy_name}
                        </Text>
                      </Link>
                    ),
                    sortValue: ({ policy_name }) => policy_name,
                    maxWidth: 200,
                  },

                  {
                    label: 'Violation Time',
                    value: ({ creationTimestamp }) => (
                      <Timestamp time={creationTimestamp} />
                    ),
                    defaultSort: true,
                    sortValue: ({ creationTimestamp }) => {
                      const t =
                        creationTimestamp &&
                        new Date(creationTimestamp).getTime();
                      return t * -1;
                    },
                  },
                ]}
                hideSearchAndFilters
                onColumnHeaderClick={columnHeaderHandler(
                  queryState,
                  setQueryState,
                )}
              />
            </TableWrapper>
            <FilterDrawer
              onClose={() => setFilterDrawerOpen(false)}
              open={filterDrawerOpen}
            >
              <QueryInput />

              <Filters facets={filteredFacets || []} />
            </FilterDrawer>
          </Flex>
          <PaginationControls
            queryState={queryState}
            setQueryState={setQueryState}
            count={data?.objects?.length || 0}
          />
        </>
      )}
    </QueryStateProvider>
  ) : (
    <NotificationsWrapper>
      <Spacer padding="medium" />
      <Flex center align>
        <MessageBox>
          <Spacer padding="base" />
          <Text size="large" semiBold>
            Explorer Disabled
          </Text>
          <Spacer padding="small" />
          <Text size="medium" capitalize>
            the explorer service is disabled and it's required to view the audit
            logs.
          </Text>
          <Spacer padding="small" />
          <Flex wide align center>
            <LinkTag
              href="https://docs.gitops.weave.works/docs/explorer/configuration/"
              newTab
            >
              <Button id="navigate-to-imageautomation">
                EXPLORER CONFIGRATION GUIDE
              </Button>
            </LinkTag>
          </Flex>
        </MessageBox>
      </Flex>
    </NotificationsWrapper>
  );
};
export default PolicyAuditList;
