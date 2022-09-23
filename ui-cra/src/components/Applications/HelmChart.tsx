import { FC } from 'react';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { HelmChartDetail, Kind, useGetObject } from '@weaveworks/weave-gitops';
import { HelmChart } from '@weaveworks/weave-gitops/ui/lib/objects';

type Props = {
  name: string;
  namespace: string;
  clusterName: string;
};

const WGApplicationsHelmChart: FC<Props> = props => {
  const { name, namespace, clusterName } = props;
  const {
    data: helmChart,
    isLoading,
    error,
  } = useGetObject<HelmChart>(name, namespace, Kind.HelmChart, clusterName);

  return (
    <PageTemplate documentTitle="WeGO · Helm Chart">
      <SectionHeader
        path={[
          {
            label: 'Applications',
            url: '/applications',
          },
          {
            label: 'Sources',
            url: '/sources',
          },
          {
            label: `${props.name}`,
          },
        ]}
      />
      <ContentWrapper
        loading={isLoading}
        errors={
          error ? [{ clusterName, namespace, message: error?.message }] : []
        }
      >
        <HelmChartDetail helmChart={helmChart} {...props} />
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsHelmChart;
