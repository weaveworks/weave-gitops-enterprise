import { Flex, Link, Text, formatURL } from '@weaveworks/weave-gitops';
import moment from 'moment';
import { GetPolicyConfigResponse } from '../../../cluster-services/cluster_services.pb';
import { Routes, getKindRoute } from '../../../utils/nav';
import { RowHeaders, SectionRowHeader } from '../../RowHeader';
import { TargetItemKind } from '../PolicyConfigStyles';

function PolicyConfigHeaderSection({
  age,
  clusterName,
  match = {},
  matchType,
}: GetPolicyConfigResponse) {
  const defaultHeaders: Array<SectionRowHeader> = [
    {
      rowkey: 'Cluster',
      value: clusterName,
    },
    {
      rowkey: 'Age',
      value: moment(age).fromNow(),
    },
  ];

  const target: any[] = [];
  Object.entries(match).forEach(([key, val]) => {
    if (key === matchType) target.push(...val);
  });

  const getMatchedItem = (
    item: any,
    clusterName: string | undefined,
    type: string,
  ) => {
    switch (type) {
      case 'apps':
        return (
          <Flex key={item.name}>
            {item.namespace === '' ? (
              <span data-testid={`matchItem${item.name}`}>*/{item.name}</span>
            ) : (
              <Link
                to={formatURL(getKindRoute(item.kind), {
                  clusterName: clusterName,
                  name: item.name,
                  namespace: item.namespace || null,
                })}
              >
                <span data-testid={`matchItem${item.name}`}>
                  {item.namespace}/{item.name}
                </span>
              </Link>
            )}
            <TargetItemKind data-testid={`matchItemKind${item.kind}`}>
              {item.kind}
            </TargetItemKind>
          </Flex>
        );
      case 'resources':
        return (
          <Flex key={item.name}>
            <span data-testid={`matchItem${item.name}`}>
              {item.namespace === '' ? '*' : item.namespace}/{item.name}
            </span>
            <TargetItemKind data-testid={`matchItemKind${item.kind}`}>
              {item.kind}
            </TargetItemKind>
          </Flex>
        );
      case 'workspaces':
        return (
          <Flex key={item}>
            <Link
              to={formatURL(Routes.WorkspaceDetails, {
                clusterName: clusterName,
                workspaceName: item,
              })}
            >
              <span data-testid={`matchItem${item}`}>{item}</span>
            </Link>
          </Flex>
        );
      case 'namespaces':
        return (
          <li key={item} data-testid={`matchItem${item}`}>
            {item}
          </li>
        );
    }
  };

  return (
    <Flex column gap="32">
      <Flex wide column gap="8">
        <RowHeaders rows={defaultHeaders} />
      </Flex>

      <Flex column gap="16">
        <Text capitalize semiBold size="medium">
          Applied To
        </Text>

        <Flex column gap="8">
          <Text capitalize>
            {matchType} ({target?.length})
          </Text>

          <Flex start column gap="8">
            {target?.map((item: any) =>
              getMatchedItem(item, clusterName, matchType || ''),
            )}
          </Flex>
        </Flex>
      </Flex>
    </Flex>
  );
}

export default PolicyConfigHeaderSection;
