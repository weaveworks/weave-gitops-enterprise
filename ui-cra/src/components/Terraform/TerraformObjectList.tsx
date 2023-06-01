import styled from 'styled-components';
import { useListTerraformObjects } from '../../contexts/Terraform';
import { NotificationsWrapper } from '../Layout/NotificationsWrapper';
import TerraformListTable from './TerraformListTable';
import { Page } from '@weaveworks/weave-gitops';
type Props = {
  className?: string;
};

function TerraformObjectList({ className }: Props) {
  const { isLoading, data, error } = useListTerraformObjects();

  return (
    <Page
      loading={isLoading}
      path={[
        {
          label: 'Terraform Objects',
        },
      ]}
    >
      <NotificationsWrapper errors={error ? [error] : data?.errors || []}>
        <TerraformListTable rows={data?.objects || []} />
      </NotificationsWrapper>
    </Page>
  );
}

export default styled(TerraformObjectList).attrs({
  className: TerraformObjectList.name,
})``;
