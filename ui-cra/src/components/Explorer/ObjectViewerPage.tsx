import styled from 'styled-components';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';
import GenericObjectViewer, { ObjectViewerProps } from './GenericObjectViewer';

function ObjectViewerPage({ className, ...props }: ObjectViewerProps) {
  return (
    <PageTemplate
      documentTitle="Explorer"
      path={[
        { label: 'Explorer', url: '/explorer' },
        {
          label: `${props.kind}/${props.name}`,
        },
      ]}
    >
      <ContentWrapper>
        <GenericObjectViewer {...props} />
      </ContentWrapper>
    </PageTemplate>
  );
}

export default styled(ObjectViewerPage).attrs({
  className: ObjectViewerPage.name,
})``;
