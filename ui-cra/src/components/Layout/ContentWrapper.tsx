import { FC } from 'react';
import styled, { css } from 'styled-components';
import { spacing } from 'weaveworks-ui-components/lib/theme/selectors';

export const pageDimensionsCss = css`
  width: 100%;
`;

export const PageWrapper = styled.div`
  ${pageDimensionsCss}
  margin: 0 auto;
`;

export const contentCss = css`
  padding: ${spacing('large')};
  margin: ${spacing('small')};
  background-color: white;
  border-radius: ${spacing('xs')};
`;

export const Content = styled.div`
  ${contentCss}
`;

const HelpLinkWrapper = styled.div`
  display: flex;
  justify-content: flex-end;
  margin-top: ${spacing('large')};
  color: ${({ theme }) => theme.colors.gray600};
  white-space: nowrap;
  line-height: 1.5em;
  a {
    color: ${({ theme }) => theme.colors.blue600};
  }
`;

export const ContentWrapper: FC = ({ children }) => {
  return (
    <Content>
      {children}
      <HelpLinkWrapper>
        Need help? Contact us at&nbsp;
        <a href="mailto:support@weave.works">support@weave.works</a>
      </HelpLinkWrapper>
    </Content>
  );
};
