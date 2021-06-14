import React, { FC } from 'react';
import styled from 'styled-components';
import { contentDimensionsCss } from './ContentWrapper';
import { spacing } from 'weaveworks-ui-components/lib/theme/selectors';

const Wrapper = styled.div`
  ${contentDimensionsCss}
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  margin: ${spacing('small')} auto;
`;

const HelpLinkWrapper = styled.div`
  color: ${({ theme }) => theme.colors.gray600};
  white-space: nowrap;
  line-height: 1.5em;
  a {
    color: ${({ theme }) => theme.colors.blue600};
  }
`;

export const Footer: FC = () => (
  <Wrapper>
    <HelpLinkWrapper>
      Need help? Contact us at{' '}
      <a href="mailto:support@weave.works">support@weave.works</a>
    </HelpLinkWrapper>
  </Wrapper>
);
