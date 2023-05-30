import ViolationDetails from './ViolationDetails';
import { useGetPolicyValidationDetails } from '../../../contexts/PolicyViolations';
import { Breadcrumb } from '../../Breadcrumbs';
import { Routes } from '../../../utils/nav';
import { Page, formatURL } from '@weaveworks/weave-gitops';

const PolicyViolationDetails = ({
  id,
  clusterName,
  source,
  sourcePath,
}: {
  id: string;
  clusterName: string;
  source?: string;
  sourcePath?: string;
}) => {
  const { data, isLoading } = useGetPolicyValidationDetails({
    clusterName,
    violationId: id,
  });
  const headerPath: Breadcrumb[] = !!source
    ? [
        { label: 'Applications', url: Routes.Applications },
        {
          label: data?.violation?.entity || '',
          url: formatURL(`/${sourcePath}/violations`, {
            name: data?.violation?.entity,
            namespace: data?.violation?.namespace,
            clusterName: clusterName,
          }),
        },
        { label: data?.violation?.message || '' },
      ]
    : [
        { label: 'Clusters', url: Routes.Clusters },
        {
          label: 'Violation Logs',
          url: Routes.PolicyViolations,
        },
        { label: data?.violation?.name || '' },
      ];
  return (
    <Page loading={isLoading} path={headerPath}>
      {data?.violation && (
        <ViolationDetails violation={data.violation} source={source} />
      )}
    </Page>
  );
};

export default PolicyViolationDetails;
