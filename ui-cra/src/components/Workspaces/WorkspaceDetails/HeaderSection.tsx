import { Workspace } from '../../../cluster-services/cluster_services.pb';
import { generateRowHeaders, SectionRowHeader } from '../../RowHeader';

function HeaderSection({ name, namespaces }: Workspace) {
  const defaultHeaders: Array<SectionRowHeader> = [
    {
      rowkey: 'Tenant Name',
      value: name,
    },
    {
      rowkey: 'Namespaces',
      value: namespaces?.join(', '),
    },
  ];

  return <>{generateRowHeaders(defaultHeaders)}</>;
}

export default HeaderSection;
