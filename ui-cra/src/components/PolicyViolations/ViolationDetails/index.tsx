import { PageTemplate } from '../../Layout/PageTemplate';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import ViolationDetails from './ViolationDetails';
import { useGetPolicyValidationDetails } from '../../../contexts/PolicyViolations';
import { Breadcrumb } from '../../Breadcrumbs';

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
  const { data, error, isLoading } = useGetPolicyValidationDetails({
    clusterName,
    violationId: id,
  });
  const headerPath: Breadcrumb[] = !!source
    ? [
        { label: 'Applications', url: '/applications' },
        {
          label: data?.violation?.entity || '',
          url: `/${sourcePath}/violations?clusterName=${clusterName}&name=${data?.violation?.entity}&namespace=${data?.violation?.namespace}`,
        },
        { label: data?.violation?.message || '' },
      ]
    : [
        { label: 'Clusters', url: '/clusters' },
        {
          label: 'Violation Logs',
          url: '/clusters/violations',
        },
        { label: data?.violation?.name || '' },
      ];
  return (
    <PageTemplate documentTitle="WeGO · Violation Logs" path={headerPath}>
      <ContentWrapper loading={isLoading} errorMessage={error?.message}>
        {data?.violation && (
          <ViolationDetails violation={data.violation} source={source} />
        )}
      </ContentWrapper>
    </PageTemplate>
  );
};

export default PolicyViolationDetails;
