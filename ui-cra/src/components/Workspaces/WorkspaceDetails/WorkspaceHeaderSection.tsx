import { Button, Flex, Icon, IconType } from '@weaveworks/weave-gitops';
import { useHistory } from 'react-router-dom';
import { Workspace } from '../../../cluster-services/cluster_services.pb';
import { toFilterQueryString } from '../../../utils/FilterQueryString';
import RowHeader from '../../RowHeader';

function WorkspaceHeaderSection({ name, namespaces, clusterName }: Workspace) {
  const history = useHistory();

  return (
    <Flex column gap="16">
      <Button
        onClick={() => {
          const filtersValues = toFilterQueryString([
            { key: 'tenant', value: name || '' },
            { key: 'clusterName', value: clusterName || '' },
          ]);
          history.push(`/applications?filters=${filtersValues}`);
        }}
      >
        <Icon type={IconType.FilterIcon} color="primary10" size="small" />
        GO TO TENANT APPLICATIONS
      </Button>
      <Flex column gap="8">
        <RowHeader rowkey="Workspace Name" value={name} />
        <RowHeader rowkey="Namespaces" value={namespaces?.join(', ')} />
      </Flex>
    </Flex>
  );
}

export default WorkspaceHeaderSection;
