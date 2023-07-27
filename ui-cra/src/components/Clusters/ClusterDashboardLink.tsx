import { formatURL } from '@weaveworks/weave-gitops';
import { Link } from 'react-router-dom';
import { Routes } from '../../utils/nav';

interface Props {
  clusterName: string;
  namespace?: string;
}

export function formatClusterDashboardUrl({ clusterName, namespace }: Props) {
  if (clusterName === 'management') return '';
  // clusterName includes namespace in manay places across the console
  // Taking in consideration that cluster Name doesn't contain any / separator

  if (!clusterName) {
    // https://github.com/weaveworks/weave-gitops-enterprise/issues/2332
    return '';
  }
  let ns = '';
  let cl = '';
  const cls = clusterName?.split('/');

  if (cls.length > 1) {
    [ns, cl] = cls;
  } else {
    cl = clusterName;
  }

  return formatURL(Routes.ClusterDashboard, {
    clusterName: cl,
    namespace: namespace || ns,
  });
}

export function ClusterDashboardLink({
  clusterName,
  namespace,
}: Props): JSX.Element {
  const clsUrl = formatClusterDashboardUrl({ clusterName, namespace });
  return <>{clsUrl ? <Link to={clsUrl}>{clusterName}</Link> : clusterName}</>;
}
