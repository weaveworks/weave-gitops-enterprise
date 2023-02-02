import {
  SourceLink,
  useGetObject,
  Kind,
  V2Routes,
  Page,
  Metadata,
} from '@weaveworks/weave-gitops';
import { InfoField } from '@weaveworks/weave-gitops/ui/components/InfoList';
import { ListError } from '@weaveworks/weave-gitops/ui/lib/api/core/core.pb';
import { FluxObject } from '@weaveworks/weave-gitops/ui/lib/objects';
import * as React from 'react';
import styled from 'styled-components';
import { Routes } from '../../../utils/nav';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';

import ImageAutomationDetails from '../ImageAutomationDetails';

type Props = {
  className?: string;
  name: string;
  namespace: string;
  clusterName: string;
};
function getInfoList(data: FluxObject, clusterName: string): InfoField[] {
  const {
    kind,
    spec: { update, git, sourceRef },
  } = data.obj;
  const { path } = update;
  const { commit, checkout, push } = git;

  return [
    ['Kind', kind],
    ['Namespace', data.namespace],
    ['Source', <SourceLink sourceRef={sourceRef} clusterName={clusterName} />],
    ['Update Path', path],
    ['Checkout Branch', checkout?.ref?.branch],
    ['Push Branch', push.branch],
    ['Author Name', commit.author?.name],
    ['Author Email', commit.author?.email],
    ['Commit Template', <code> {commit.messageTemplate}</code>],
  ];
}
const kind = 'ImageUpdateAutomation' as Kind; //Kind.ImageUpdateAutomation
function ImageAutomationUpdatesDetails({
  className,
  name,
  namespace,
  clusterName,
}: Props) {
  const { data, isLoading, error } = useGetObject<FluxObject>(
    name,
    namespace,
    kind, //Kind.ImageUpdateAutomation,
    clusterName,
    {
      refetchInterval: 5000,
    },
  );

  const rootPath = Routes.ImageAutomationUpdatesDetails;
  return (
    <PageTemplate
      documentTitle="Image Automation Updates"
      path={[
        { label: 'Image Automation', url: Routes.ImageAutomation },
        { label: name },
      ]}
    >
      <ContentWrapper loading={isLoading} errors={[error as ListError]}>
        {!!data && (
          <ImageAutomationDetails
            data={data}
            kind={kind}
            infoFields={getInfoList(data, data.clusterName)}
            rootPath={rootPath}
          >
            <Metadata metadata={data.metadata} labels={data.labels} />
          </ImageAutomationDetails>
        )}
      </ContentWrapper>
    </PageTemplate>
  );
}

export default styled(ImageAutomationUpdatesDetails)``;
