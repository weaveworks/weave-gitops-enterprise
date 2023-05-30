import {
  Kind,
  Metadata,
  useGetObject,
  V2Routes,
  Page,
} from '@weaveworks/weave-gitops';
import { ImagePolicy } from '@weaveworks/weave-gitops/ui/lib/objects';
import styled from 'styled-components';
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
      {data && (
        <ImageAutomationDetails
          data={data}
          kind={Kind.ImagePolicy}
          infoFields={[
            ['Kind', Kind.ImagePolicy],
            ['Namespace', data?.namespace],
            ['Namespace', data?.clusterName],
            ['Image Policy', data?.imagePolicy?.type || ''],
            ['Order/Range', data?.imagePolicy?.value],
          ]}
          rootPath={rootPath}
        >
          <Metadata metadata={data?.metadata} labels={data?.labels} />
        </ImageAutomationDetails>
      )}
    </Page>
  );
}

export default styled(ImagePolicyDetails)``;
