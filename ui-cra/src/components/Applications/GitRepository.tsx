import { FC } from 'react';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';
import {
  GitRepositoryDetail,
  Kind,
  useGetObject,
  V2Routes,
} from '@weaveworks/weave-gitops';
import { GitRepository } from '@weaveworks/weave-gitops/ui/lib/objects';
import { EditButton } from '../EditButton';
import { Routes } from '../../utils/nav';

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
    <PageTemplate
      documentTitle="Git Repository"
      path={[
        {
          label: 'Applications',
          url: Routes.Applications,
        },
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
        <GitRepositoryDetail
          gitRepository={gitRepository}
          customActions={[
            <EditButton resource={gitRepository} isLoading={isLoading} />,
          ]}
          {...props}
        />
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsGitRepository;
