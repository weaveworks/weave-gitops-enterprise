import { formatURL, Link } from '@weaveworks/weave-gitops';
import moment from 'moment';
import { useEffect, useState } from 'react';
import { GetPolicyConfigResponse } from '../../../cluster-services/cluster_services.pb';
import { generateRowHeaders, SectionRowHeader } from '../../RowHeader';
import { usePolicyConfigStyle } from '../PolicyConfigStyles';
import { getKindRoute, Routes } from '../../../utils/nav';

function PolicyConfigHeaderSection({
  age,
  clusterName,
  match = {},
  matchType,
}: GetPolicyConfigResponse) {
  const classes = usePolicyConfigStyle();
  const [target, setTarget] = useState<any[]>();
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

  useEffect(() => {
    const matchTarget = Object.entries(match)
      .filter(item => item[0] === matchType)
      .map(item => item[1]);
    setTarget(matchTarget.flat());
  }, [matchType, match]);
  
  const getMatchedItem = (
    item: any,
    clusterName: string | undefined,
    type: string,
  ) => {
    switch (type) {
      case 'apps':
        return (
          <li key={`${item.name}`}>
            <Link
              to={formatURL(getKindRoute(item.kind), {
                clusterName: clusterName,
                name: item.name,
                namespace: item.namespace || null,
              })}
            >
              <span>
                {item.namespace === '' ? <span>*</span> : item.namespace}/
                {item.name}
              </span>
            </Link>
            <span className={`${classes.targetItemKind} ${classes.capitlize}`}>
              {item.kind}
            </span>
          </li>
        );
      case 'resources':
        return (
          <li key={`${item.name}`}>
            <span>
              {item.namespace === '' ? <span>*</span> : item.namespace}/
              {item.name}
            </span>
            <span className={`${classes.targetItemKind} ${classes.capitlize}`}>
              {item.kind}
            </span>
          </li>
        );
      case 'workspaces':
        return (
          <li key={item}>
            <Link
              to={formatURL(Routes.WorkspaceDetails, {
                clusterName: clusterName,
                workspaceName: item.name,
              })}
            >
              {item}
            </Link>
          </li>
        );
      case 'namespaces':
        return <li key={item}>{item}</li>;
    }
  };

  return (
    <div>
      {generateRowHeaders(defaultHeaders)}
      <div>
        <label className={classes.sectionTitle}>Applied To</label>
        <div
          className={`${classes.sectionTitle} ${classes.capitlize}`}
          style={{ fontWeight: 'normal', marginTop: '12px' }}
        >
          {matchType}

          <span> ({target?.length})</span>
        </div>
        <ul className={classes.targetItemsList}>
          {target?.map((item: any) =>
            getMatchedItem(item, clusterName, matchType || ''),
          )}
        </ul>
      </div>
    </div>
  );
}

export default PolicyConfigHeaderSection;
