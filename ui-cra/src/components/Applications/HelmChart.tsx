import { FC } from 'react';
import {
  HelmChartDetail,
  Kind,
  Page,
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
    <Page
      loading={isLoading}
      error={error ? [{ clusterName, namespace, message: error?.message }] : []}
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
      <HelmChartDetail
        helmChart={helmChart}
        customActions={[<EditButton resource={helmChart} />]}
        {...props}
      />
    </Page>
  );
};

export default WGApplicationsHelmChart;
