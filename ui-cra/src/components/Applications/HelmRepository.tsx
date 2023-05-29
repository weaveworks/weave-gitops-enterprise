import { FC } from 'react';
import { ContentWrapper } from '../Layout/ContentWrapper';
import {
  HelmRepositoryDetail,
  Kind,
  Page,
  useGetObject,
  V2Routes,
} from '@weaveworks/weave-gitops';
import { HelmRepository } from '@weaveworks/weave-gitops/ui/lib/objects';
import { EditButton } from '../Templates/Edit/EditButton';

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
    <Page
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
        <HelmRepositoryDetail
          helmRepository={helmRepository}
          customActions={[<EditButton resource={helmRepository} />]}
          {...props}
        />
      </ContentWrapper>
    </Page>
  );
};

export default WGApplicationsHelmRepository;
