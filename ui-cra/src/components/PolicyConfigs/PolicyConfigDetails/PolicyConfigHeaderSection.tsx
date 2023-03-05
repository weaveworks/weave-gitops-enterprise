import moment from 'moment';
import { useEffect, useState } from 'react';
import { GetPolicyConfigResponse } from '../../../cluster-services/cluster_services.pb';
import { generateRowHeaders, SectionRowHeader } from '../../RowHeader';
import { usePolicyConfigStyle } from '../PolicyConfigStyles';
interface Target {
  targetName: string;
  targetList: any[];
}
function PolicyConfigHeaderSection({
  age,
  clusterName,
  match,
}: GetPolicyConfigResponse) {
  const classes = usePolicyConfigStyle();
  const [target, setTarget] = useState<Target>();
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
    for (var targett in match) {
      const m = match[targett as keyof typeof match];
      if (m?.length) setTarget({ targetName: targett, targetList: m });
    }
  }, []);
  return (
    <div>
      {generateRowHeaders(defaultHeaders)}
      <div>
        <label className={classes.sectionTitle}>Applied To</label>
        <div
          className={`${classes.sectionTitle} ${classes.capitlize}`}
          style={{ fontWeight: 'normal', marginTop: '12px' }}
        >
          {target?.targetName}
          <span> ({target?.targetList.length})</span>
        </div>
        <ul className={classes.targetItemsList}>
          {target?.targetName === 'resources' || target?.targetName === 'apps'
            ? target?.targetList.map((item: any) => (
                <li key={`${item.name}`}>
                  <span>
                    {item.namespace}/{item.name}
                  </span>
                  <span
                    className={`${classes.targetItemKind} ${classes.capitlize}`}
                  >
                    {item.kind}
                  </span>
                </li>
              ))
            : target?.targetList.map((item: any) => (
                <li key={item}>{item}</li>
              ))}
        </ul>
      </div>
    </div>
  );
}

export default PolicyConfigHeaderSection;
