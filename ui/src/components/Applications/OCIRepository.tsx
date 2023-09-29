import { OCIRepository } from '@weaveworks/weave-gitops/ui/lib/objects';
import {
  OCIRepositoryDetail,
  Kind,
  useGetObject,
  V2Routes,
} from '@weaveworks/weave-gitops';
import { FC } from 'react';
import { Page } from '../Layout/App';
import { NotificationsWrapper } from '../Layout/NotificationsWrapper';
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
      <NotificationsWrapper
        errors={
          error ? [{ clusterName, namespace, message: error?.message }] : []
        }
      >
        <OCIRepositoryDetail
          ociRepository={ociRepository}
          customActions={[<EditButton resource={ociRepository} />]}
          {...props}
        />
      </NotificationsWrapper>
    </Page>
  );
};

export default WGApplicationsOCIRepository;
