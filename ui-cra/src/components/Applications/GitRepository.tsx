import { FC } from 'react';
import {
  GitRepositoryDetail,
  Kind,
  Page,
  useGetObject,
  V2Routes,
} from '@weaveworks/weave-gitops';
import { GitRepository } from '@weaveworks/weave-gitops/ui/lib/objects';
import { EditButton } from '../Templates/Edit/EditButton';

type Props = {
  name: string;
  namespace: string;
  clusterName: string;
};

const WGApplicationsGitRepository: FC<Props> = props => {
  const { name, namespace, clusterName } = props;
  const {
    data: gitRepository,
    isLoading,
    error,
  } = useGetObject<GitRepository>(
    name,
    namespace,
    Kind.GitRepository,
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
      <GitRepositoryDetail
        gitRepository={gitRepository}
        customActions={[<EditButton resource={gitRepository} />]}
        {...props}
      />
    </Page>
  );
};

export default WGApplicationsGitRepository;
