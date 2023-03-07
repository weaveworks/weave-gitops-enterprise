import moment from 'moment';
import { useEffect, useState } from 'react';
import { GetPolicyConfigResponse } from '../../../cluster-services/cluster_services.pb';
import { generateRowHeaders, SectionRowHeader } from '../../RowHeader';
import { usePolicyConfigStyle } from '../PolicyConfigStyles';

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
      .filter(item => item[0] == matchType)
      .map(item => item[1]);
    setTarget(matchTarget.flat());
  }, [matchType, match]);

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
          {matchType === 'resources' || matchType === 'apps'
            ? target?.map((item: any) => (
                <li key={`${item.name}`}>
                  <span>
                    {item.namespace === '' ? <span>*</span> : item.namespace}/
                    {item.name}
                  </span>
                  <span
                    className={`${classes.targetItemKind} ${classes.capitlize}`}
                  >
                    {item.kind}
                  </span>
                </li>
              ))
            : target?.map((item: any) => <li key={item}>{item}</li>)}
        </ul>
      </div>
    </div>
  );
}

export default PolicyConfigHeaderSection;
