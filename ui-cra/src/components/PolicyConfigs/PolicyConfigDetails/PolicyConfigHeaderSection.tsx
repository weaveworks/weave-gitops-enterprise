import { formatURL, Link } from '@weaveworks/weave-gitops';
import moment from 'moment';
import { GetPolicyConfigResponse } from '../../../cluster-services/cluster_services.pb';
import { getKindRoute, Routes } from '../../../utils/nav';
import { RowHeaders, SectionRowHeader } from '../../RowHeader';
import {
  AppliedToTitle,
  SectionTitle,
  TargetItemKind,
  TargetItemsList,
} from '../PolicyConfigStyles';

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
          <li key={item.name}>
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
          </li>
        );
      case 'resources':
        return (
          <li key={item.name}>
            <span data-testid={`matchItem${item.name}`}>
              {item.namespace === '' ? '*' : item.namespace}/{item.name}
            </span>
            <TargetItemKind data-testid={`matchItemKind${item.kind}`}>
              {item.kind}
            </TargetItemKind>
          </li>
        );
      case 'workspaces':
        return (
          <li key={item}>
            <Link
              to={formatURL(Routes.WorkspaceDetails, {
                clusterName: clusterName,
                workspaceName: item,
              })}
            >
              <span data-testid={`matchItem${item}`}>{item}</span>
            </Link>
          </li>
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
    <div>
      <RowHeaders rows={defaultHeaders} />
      <div>
        <SectionTitle>Applied To</SectionTitle>
        <AppliedToTitle capitalize data-testid="appliedTo">
          <span>{matchType}</span>
          <span> ({target?.length})</span>
        </AppliedToTitle>
        <TargetItemsList>
          {target?.map((item: any) =>
            getMatchedItem(item, clusterName, matchType || ''),
          )}
        </TargetItemsList>
      </div>
    </div>
  );
}

export default PolicyConfigHeaderSection;
