import {
  Kind,
  Metadata,
  useGetObject,
  V2Routes,
} from '@weaveworks/weave-gitops';
import { ImagePolicy } from '@weaveworks/weave-gitops/ui/lib/objects';
import styled from 'styled-components';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';

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
      refetchInterval: 5000,
    },
  );
  const rootPath = V2Routes.ImagePolicyDetails;
  return (
    <PageTemplate
      documentTitle={name}
      path={[
        { label: 'Image Policies', url: V2Routes.ImagePolicies },
        { label: name },
      ]}
    >
      <ContentWrapper loading={isLoading}>
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
      </ContentWrapper>
    </PageTemplate>
  );
}

export default styled(ImagePolicyDetails)``;
