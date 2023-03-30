import { formatClusterDashboardUrl } from '../components/Clusters/ClusterDashboardLink';

// FIXME: remove this when core fixes requiring a linkResolver function
export const resolver = (path: string, params?: any) => {
  // cover cluster-dashboard resolver path
  if (path === 'ClusterDashboard') {
    const url = formatClusterDashboardUrl(params.clusterName);
    return url;
  }
  // Fix Kind as a path
  return path?.includes('/') ? path : '';
};
