import {
  Kind,
  Metadata,
  Page,
  SourceLink,
  useGetObject,
  V2Routes,
} from '@weaveworks/weave-gitops';
import { InfoField } from '@weaveworks/weave-gitops/ui/components/InfoList';
import { ImageUpdateAutomation } from '@weaveworks/weave-gitops/ui/lib/objects';
import styled from 'styled-components';
import ImageAutomationDetails from '../ImageAutomationDetails';

type Props = {
  className?: string;
  name: string;
  namespace: string;
  clusterName: string;
};
function getInfoList(
  data: ImageUpdateAutomation,
  clusterName: string,
): InfoField[] {
  const {
    kind,
    spec: { update, git },
  } = data.obj;
  const { path } = update;
  const { commit, checkout, push } = git;

  return [
    ['Kind', kind],
    ['Namespace', data?.namespace],
    ['Namespace', data?.clusterName],
    [
      'Source',
      <SourceLink sourceRef={data.sourceRef} clusterName={clusterName} />,
    ],
    ['Update Path', path],
    ['Checkout Branch', checkout?.ref?.branch],
    ['Push Branch', push?.branch],
    ['Author Name', commit.author?.name],
    ['Author Email', commit.author?.email],
    ['Commit Template', <code> {commit.messageTemplate}</code>],
  ];
}

function ImageAutomationUpdatesDetails({
  name,
  namespace,
  clusterName,
}: Props) {
  const { data, isLoading } = useGetObject<ImageUpdateAutomation>(
    name,
    namespace,
    Kind.ImageUpdateAutomation,
    clusterName,
    {
      refetchInterval: 30000,
    },
  );

  const rootPath = V2Routes.ImageAutomationUpdatesDetails;
  return (
    <Page
      loading={isLoading}
      path={[
        { label: 'Image Updates', url: V2Routes.ImageUpdates },
        { label: name },
      ]}
    >
      {data && (
        <ImageAutomationDetails
          data={data}
          kind={Kind.ImageUpdateAutomation}
          infoFields={getInfoList(data, data?.clusterName)}
          rootPath={rootPath}
        >
          <Metadata metadata={data.metadata} labels={data?.labels} />
        </ImageAutomationDetails>
      )}
    </Page>
  );
}

export default styled(ImageAutomationUpdatesDetails)``;
