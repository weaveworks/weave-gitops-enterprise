import {
  AppContext,
  filterByStatusCallback,
  filterConfig,
  ReconciledObjectsAutomation,
  RequestStateHandler,
} from '@weaveworks/weave-gitops';
import * as React from 'react';
import styled from 'styled-components';
import FluxObjectsTable from './FluxObjectsTable';

interface Props {
  className?: string;
  reconciledObjectsAutomation: ReconciledObjectsAutomation;
}

function ReconciledObjectsTable({
  className,
  reconciledObjectsAutomation,
}: Props) {
  const { objects, isLoading, error } = reconciledObjectsAutomation;

  const initialFilterState = {
    ...filterConfig(objects, 'type'),
    ...filterConfig(objects, 'namespace'),
    ...filterConfig(objects, 'status', filterByStatusCallback),
  };

  const { setDetailModal } = React.useContext(AppContext);

  // console.log(setDetailModal);

  // console.log(objects);

  return (
    // @ts-ignore
    <RequestStateHandler loading={isLoading} error={error}>
      <FluxObjectsTable
        className={className}
        //@ts-ignore
        objects={objects}
        onClick={setDetailModal}
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
