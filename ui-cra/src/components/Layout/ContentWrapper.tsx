import styled, { css } from 'styled-components';
import { spacing } from 'weaveworks-ui-components/lib/theme/selectors';

export const pageDimensionsCss = css`
  width: 100%;
`;

export const PageWrapper = styled.div`
  ${pageDimensionsCss}
  margin: 0 auto;
`;

export const contentDimensionsCss = css`
  padding: ${spacing('large')};
`;

export const ContentWrapper = styled.div`
  ${contentDimensionsCss}
  margin: ${spacing('small')};
  background-color: white;
  border-radius: ${spacing('xs')};
`;
