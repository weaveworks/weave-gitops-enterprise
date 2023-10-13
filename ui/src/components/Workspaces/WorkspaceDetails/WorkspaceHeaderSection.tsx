import { Button, Flex, Icon, IconType } from '@weaveworks/weave-gitops';
import { useHistory } from 'react-router-dom';
import styled from 'styled-components';
import { Workspace } from '../../../cluster-services/cluster_services.pb';
import { toFilterQueryString } from '../../../utils/FilterQueryString';
import RowHeader from '../../RowHeader';

const Header = styled(Flex)`
  margin-bottom: ${props => props.theme.spacing.medium};
`;

function WorkspaceHeaderSection({ name, namespaces, clusterName }: Workspace) {
  const history = useHistory();

  return (
    <Flex column gap="16">
      <Button
        startIcon={<Icon type={IconType.FilterIcon} size="base" />}
        onClick={() => {
          const filtersValues = toFilterQueryString([
            { key: 'tenant', value: name || '' },
            { key: 'clusterName', value: clusterName || '' },
          ]);
          history.push(`/applications?filters=${filtersValues}`);
        }}
      >
        GO TO TENANT APPLICATIONS
      </Button>
      <Header column gap="8">
        <RowHeader rowkey="Workspace Name" value={name} />
        <RowHeader rowkey="Namespaces" value={namespaces?.join(', ')} />
      </Header>
    </Flex>
  );
}

export default WorkspaceHeaderSection;
