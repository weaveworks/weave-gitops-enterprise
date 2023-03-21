import { DataTable } from '@weaveworks/weave-gitops';
import _ from 'lodash';
import styled from 'styled-components';
import { AccessRule } from '../../api/query/query.pb';
import { useListAccessRules } from '../../hooks/query';

type Props = {
  className?: string;
};

function AccessRulesDebugger({ className }: Props) {
  const { data: rules } = useListAccessRules();
  return (
    <div className={className}>
      <DataTable
        fields={[
          { label: 'Cluster', value: 'cluster' },
          {
            label: 'Subjects',
            value: (r: AccessRule) =>
              _.map(r.subjects, 'name').join(', ') || null,
          },
          {
            label: 'Accessible Kinds',
            value: (r: AccessRule) => r?.accessibleKinds?.join(', ') || null,
          },
        ]}
        rows={rules?.rules}
      />
    </div>
  );
}

export default styled(AccessRulesDebugger).attrs({
  className: AccessRulesDebugger.name,
})``;
