import {
  Kind,
  Metadata,
  useGetObject,
  V2Routes,
} from '@weaveworks/weave-gitops';
import { ImagePolicy } from '@weaveworks/weave-gitops/ui/lib/objects';
import styled from 'styled-components';
import { Page } from '../../Layout/App';
import { NotificationsWrapper } from '../../Layout/NotificationsWrapper';
import ImageAutomationDetails from '../ImageAutomationDetails';

type Props = {
  className?: string;
  name: string;
  namespace: string;
  clusterName: string;
};

function ImagePolicyDetails({ name, namespace, clusterName }: Props) {
  const { data, isLoading } = useGetObject<ImagePolicy>(
    name,
    namespace,
    Kind.ImagePolicy,
    clusterName,
    {
      refetchInterval: 30000,
    },
  );
  const rootPath = V2Routes.ImagePolicyDetails;

  return (
    <Page
      loading={isLoading}
      path={[
        { label: 'Image Policies', url: V2Routes.ImagePolicies },
        { label: name },
      ]}
    >
      <NotificationsWrapper>
        {data && (
          <ImageAutomationDetails
            data={data}
            kind={Kind.ImagePolicy}
            infoFields={[
              ['Kind', Kind.ImagePolicy],
              ['Namespace', data?.namespace],
              ['Cluster', data?.clusterName],
              ['Image Policy', data?.imagePolicy?.type || ''],
              ['Order/Range', data?.imagePolicy?.value],
            ]}
            rootPath={rootPath}
          >
            <Metadata metadata={data?.metadata} labels={data?.labels} />
          </ImageAutomationDetails>
        )}
      </NotificationsWrapper>
    </Page>
  );
}

export default styled(ImagePolicyDetails)``;
