import {
  AppContext,
  filterByStatusCallback,
  filterConfig,
  FluxObject,
  FluxObjectsTable,
  RequestStateHandler,
} from '@weaveworks/weave-gitops';
import * as React from 'react';
import styled from 'styled-components';
import {
  GitOpsSet,
  GroupVersionKind,
  ResourceRef,
} from '../../api/gitopssets/types.pb';
import { useGetReconciledTree } from '../../hooks/gitopssets';
import { RequestError } from '../../types/custom';

interface ReconciledVisualizationProps {
  className?: string;
  gitOpsSet: GitOpsSet;
}

export const getInventory = (gs: GitOpsSet) => {
  const entries = gs?.inventory || [];
  return Array.from(
    new Set(
      entries.map((entry: ResourceRef) => {
        // entry is namespace_name_group_kind, but name can contain '_' itself
        const parts = entry?.id?.split('_');
        const kind = parts?.[parts.length - 1];
        const group = parts?.[parts.length - 2];
        return { group, version: entry.version, kind };
      }),
    ),
  );
};

function ReconciledObjectsTable({
  className,
  gitOpsSet,
}: ReconciledVisualizationProps) {
  const {
    data: objs,
    error,
    isLoading,
  } = useGetReconciledTree(
    gitOpsSet.name || '',
    gitOpsSet.namespace || '',
    'GitOpsSet',
    getInventory(gitOpsSet) as GroupVersionKind[],
    gitOpsSet.clusterName,
  );

  const initialFilterState = {
    ...filterConfig(objs, 'type'),
    ...filterConfig(objs, 'namespace'),
    ...filterConfig(objs, 'status', filterByStatusCallback),
  };

  const { setNodeYaml } = React.useContext(AppContext);

  return (
    <RequestStateHandler loading={isLoading} error={error as RequestError}>
      <FluxObjectsTable
        className={className}
        objects={objs as FluxObject[]}
        onClick={setNodeYaml}
        initialFilterState={initialFilterState}
      />
    </RequestStateHandler>
  );
}
export default styled(ReconciledObjectsTable).attrs({
  className: ReconciledObjectsTable.name,
})`
  td:nth-child(5),
  td:nth-child(6) {
    white-space: pre-wrap;
  }
  td:nth-child(5) {
    overflow-wrap: break-word;
    word-wrap: break-word;
  }
`;
