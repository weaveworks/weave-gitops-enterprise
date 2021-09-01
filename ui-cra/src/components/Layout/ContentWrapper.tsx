import { FC } from 'react';
import styled, { css } from 'styled-components';
import { spacing } from 'weaveworks-ui-components/lib/theme/selectors';
import useVersions from '../../contexts/Versions';

export const pageDimensionsCss = css`
  width: 100%;
`;

export const PageWrapper = styled.div`
  ${pageDimensionsCss}
  margin: 0 auto;
`;

const small = spacing('small');
const medium = spacing('medium');
const large = spacing('large');

export const contentCss = css`
  margin: ${medium} ${small} 0 ${small};
  padding: ${large} ${medium} ${medium} ${medium};
  background-color: white;
  border-radius: ${spacing('xs')};
`;

export const Content = styled.div`
  ${contentCss}
`;

const HelpLinkWrapper = styled.div`
  margin: 0 ${small};
  padding: ${small} ${small};
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  color: ${({ theme }) => theme.colors.gray600};
  white-space: nowrap;
  line-height: 1.5em;
  a {
    color: ${({ theme }) => theme.colors.blue600};
  }
`;

export const ContentWrapper: FC = ({ children }) => {
  const versions = useVersions();
  return (
    <>
      <Content>{children}</Content>
      <HelpLinkWrapper>
        <div>
          Need help? Contact us at{' '}
          <a href="mailto:support@weave.works">support@weave.works</a>
        </div>
        <div>Version {versions.versions?.capiServer}</div>
      </HelpLinkWrapper>
    </>
  );
};
