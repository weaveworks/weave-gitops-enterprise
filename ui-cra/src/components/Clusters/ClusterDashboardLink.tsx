import { formatURL } from '@weaveworks/weave-gitops';
import { Link } from 'react-router-dom';
import { Routes } from '../../utils/nav';

export function formatClusterDashboardUrl(clusterName: string): string {
  if (clusterName === 'management') return '';
  // clusterName includes namespace in manay places across the console
  // Taking in consideration that cluster Name doesn't contain any / separator

  if (!clusterName) {
    // https://github.com/weaveworks/weave-gitops-enterprise/issues/2332
    return '';
  }

  const cls = clusterName?.split('/');
  let url = cls[0];
  if (cls.length > 1) {
    url = cls[1];
  }
  return formatURL(Routes.ClusterDashboard, {
    clusterName: url,
  });
}

export function ClusterDashboardLink({ clusterName }: { clusterName: string }) {
  const clsUrl = formatClusterDashboardUrl(clusterName || '');
  return <>{clsUrl ? <Link to={clsUrl}>{clusterName}</Link> : clusterName}</>;
}
