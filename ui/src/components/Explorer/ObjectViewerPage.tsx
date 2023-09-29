import { Page } from '@weaveworks/weave-gitops';
import styled from 'styled-components';
import GenericObjectViewer, { ObjectViewerProps } from './GenericObjectViewer';

function ObjectViewerPage({ className, ...props }: ObjectViewerProps) {
  return (
    <Page
      path={[
        { label: 'Explorer', url: '/explorer' },
        {
          label: `${props.kind}/${props.name}`,
        },
      ]}
    >
      <GenericObjectViewer {...props} />
    </Page>
  );
}

export default styled(ObjectViewerPage).attrs({
  className: ObjectViewerPage.name,
})``;
