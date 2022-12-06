import { formatURL } from '@weaveworks/weave-gitops';
import { Link } from 'react-router-dom';

function extractClusterName(cluster: string, includeNamespace = true): string {
  if (cluster === 'management') return '';
  if (includeNamespace) {
    return cluster.split('/')[1];
  }
  return cluster;
}

export function ClusterDashboardLink({
  clusterName,
  namespaceIncluded = true,
  clusterDashboardRoute = '/cluster',
}: {
  clusterName: string;
  namespaceIncluded?: boolean;
  clusterDashboardRoute?: string;
}) {
  const clsName = extractClusterName(clusterName || '', namespaceIncluded);
  return (
    <>
      {clsName ? (
        <Link
          to={formatURL(clusterDashboardRoute, {
            clusterName: clsName,
          })}
        >
          {clusterName}
        </Link>
      ) : (
        clusterName
      )}
    </>
  );
}
