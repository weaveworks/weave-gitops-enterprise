import { DataTable } from '@weaveworks/weave-gitops';
import styled from 'styled-components';
import { useListAccessRules, useQueryService } from '../../hooks/query';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';

type Props = {
  className?: string;
};

function Explorer({ className }: Props) {
  const { data, error, isLoading } = useQueryService();
  const { data: rules } = useListAccessRules();

  return (
    <PageTemplate documentTitle="Explorer" path={[{ label: 'Explorer' }]}>
      <ContentWrapper>
        <div className={className}>
          <DataTable
            fields={[
              { label: 'Name', value: 'name' },
              { label: 'Kind', value: 'kind' },
              { label: 'Namespace', value: 'namespace' },
              { label: 'Cluster', value: 'cluster' },
            ]}
            rows={data?.objects}
          />
        </div>
        <br />
        <DataTable
          fields={[
            { label: 'Cluster', value: 'cluster' },
            { label: 'Principal', value: 'principal' },
            { label: 'Accessible Kinds', value: 'accessibleKinds' },
          ]}
          rows={rules?.rules}
        />
      </ContentWrapper>
    </PageTemplate>
  );
}

export default styled(Explorer).attrs({ className: Explorer.name })``;
