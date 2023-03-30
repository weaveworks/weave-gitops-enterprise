import { FC } from 'react';
import { PageTemplate } from '../Layout/PageTemplate';
import { ContentWrapper } from '../Layout/ContentWrapper';
import {
  HelmChartDetail,
  Kind,
  useGetObject,
  V2Routes,
} from '@weaveworks/weave-gitops';
import { HelmChart } from '@weaveworks/weave-gitops/ui/lib/objects';
import { EditButton } from '../Templates/Edit/EditButton';

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
    <PageTemplate
      documentTitle="Helm Chart"
      path={[
        {
          label: 'Sources',
          url: V2Routes.Sources,
        },
        {
          label: `${props.name}`,
        },
      ]}
    >
      <ContentWrapper
        loading={isLoading}
        errors={
          error ? [{ clusterName, namespace, message: error?.message }] : []
        }
      >
        <HelmChartDetail
          helmChart={helmChart}
          customActions={[<EditButton resource={helmChart} />]}
          {...props}
        />
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsHelmChart;
