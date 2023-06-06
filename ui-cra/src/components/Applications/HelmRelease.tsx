import {
  HelmReleaseDetail,
  Kind,
  useGetObject,
} from '@weaveworks/weave-gitops';
import { HelmRelease } from '@weaveworks/weave-gitops/ui/lib/objects';
import { FC } from 'react';
import { Routes } from '../../utils/nav';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';
import { EditButton } from '../Templates/Edit/EditButton';

type Props = {
  name: string;
  clusterName: string;
  namespace: string;
};

const WGApplicationsHelmRelease: FC<Props> = props => {
  const { name, namespace, clusterName } = props;
  const {
    data: helmRelease,
    isLoading,
    error,
  } = useGetObject<HelmRelease>(name, namespace, Kind.HelmRelease, clusterName);


  return (
    <PageTemplate
      documentTitle="Helm Release"
      path={[
        {
          label: 'Applications',
          url: Routes.Applications,
        },
        {
          label: `${name}`,
        },
      ]}
    >
      <ContentWrapper
        loading={isLoading}
        errors={
          error ? [{ clusterName, namespace, message: error?.message }] : []
        }
      >
        {!error && !isLoading && (
          <HelmReleaseDetail
            helmRelease={helmRelease}
            customActions={[<EditButton resource={helmRelease} />]}
            {...props}
          />
        )}
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsHelmRelease;
