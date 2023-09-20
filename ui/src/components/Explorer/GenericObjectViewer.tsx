import { Object } from '../../api/query/query.pb';
import { useQueryService } from '../../hooks/query';
import CodeView from '../CodeView';
import KeyValueTable from '../KeyValueTable';
import { LoadingPage } from '@weaveworks/weave-gitops';
import _ from 'lodash';
import styled from 'styled-components';

export type ObjectViewerProps = {
  className?: string;
  kind?: string;
  name?: string;
  namespace?: string;
  clusterName?: string;
};

function GenericObjectViewer({ className, ...rest }: ObjectViewerProps) {
  const { data, isLoading } = useQueryService({
    filters: [],
  });

  if (isLoading) {
    return <LoadingPage />;
  }

  const filters = _.pick(rest, ['kind', 'name', 'namespace']) as any;
  //   Correcting for a inconsistency with the clusterName vs cluster field.
  filters.cluster = rest.clusterName;

  const relevant = _.filter(data?.objects, filters) as Object[];
  const obj = _.first(relevant);

  return (
    <div className={className}>
      <KeyValueTable
        pairs={[
          ['Name', rest.name],
          ['Namespace', rest.namespace],
          ['Kind', rest.kind],
          ['Cluster', rest.clusterName],
          ['Status', obj?.status],
          ['Message', obj?.message],
        ]}
      />
      <CodeView
        code={JSON.stringify(JSON.parse(obj?.unstructured || '{}'), null, 2)}
      />
    </div>
  );
}

export default styled(GenericObjectViewer).attrs({
  className: GenericObjectViewer.name,
})``;
