import moment from 'moment';
import { GetPolicyConfigResponse, PolicyConfigApplicationMatch } from '../../../cluster-services/cluster_services.pb';
import { generateRowHeaders, SectionRowHeader } from '../../RowHeader';

function PolicyConfigHeaderSection({
  age,
  clusterName,
  match,
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
  const appliedTo = () => {
    for (var target in match) {
      const m = match[target as keyof typeof match];
      if (m?.length)
        return (
          <>
            <div>
              {target}
              {` ( ${m.length} )`}
            </div>
            <ul>
              {m.map((item : any)=> (
                target !== 'resources' || 'apps' ?
                <li key={`${item}`}>{item}</li>:
                <li key={`${item}`}>{item.namespace}/{item.name} <span>item.kind</span></li>
              ))}
            </ul>
          </>
        );
    }
  };
  
  return (
    <div>
      {generateRowHeaders(defaultHeaders)}
      <div>
        <label>Applied To:</label>
        {appliedTo()}
      </div>
    </div>
  );
}

export default PolicyConfigHeaderSection;
