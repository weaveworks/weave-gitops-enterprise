import { FC } from 'react';
import {
  OCIRepositoryDetail,
  Kind,
  useGetObject,
  V2Routes,
  Page,
} from '@weaveworks/weave-gitops';
import { OCIRepository } from '@weaveworks/weave-gitops/ui/lib/objects';
import { EditButton } from '../Templates/Edit/EditButton';

type Props = {
  name: string;
  namespace: string;
  clusterName: string;
};

const WGApplicationsOCIRepository: FC<Props> = props => {
  const { name, namespace, clusterName } = props;
  const {
    data: ociRepository,
    isLoading,
    error,
  } = useGetObject<OCIRepository>(
    name,
    namespace,
    Kind.OCIRepository,
    clusterName,
  );

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
      <OCIRepositoryDetail
        ociRepository={ociRepository}
        customActions={[<EditButton resource={ociRepository} />]}
        {...props}
      />
    </Page>
  );
};

export default WGApplicationsOCIRepository;
