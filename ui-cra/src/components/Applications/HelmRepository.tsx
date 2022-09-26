import { FC } from 'react';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';
import {
  HelmRepositoryDetail,
  Kind,
  useGetObject,
} from '@weaveworks/weave-gitops';
import { HelmRepository } from '@weaveworks/weave-gitops/ui/lib/objects';

type Props = {
  name: string;
  namespace: string;
  clusterName: string;
};

const WGApplicationsHelmRepository: FC<Props> = props => {
  const { name, namespace, clusterName } = props;
  const {
    data: helmRepository,
    isLoading,
    error,
  } = useGetObject<HelmRepository>(
    name,
    namespace,
    Kind.HelmRepository,
    clusterName,
  );

  return (
    <PageTemplate
      documentTitle="Helm Repository"
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
    >
      <ContentWrapper
        loading={isLoading}
        errors={
          error ? [{ clusterName, namespace, message: error?.message }] : []
        }
      >
        <HelmRepositoryDetail helmRepository={helmRepository} {...props} />
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsHelmRepository;
