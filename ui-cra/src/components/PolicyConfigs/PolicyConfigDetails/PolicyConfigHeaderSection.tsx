import moment from 'moment';
import { GetPolicyConfigResponse } from '../../../cluster-services/cluster_services.pb';
import { generateRowHeaders, SectionRowHeader } from '../../RowHeader';

function PolicyConfigHeaderSection({
  age,
  clusterName,
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

  return <>{generateRowHeaders(defaultHeaders)}</>;
}

export default PolicyConfigHeaderSection;
