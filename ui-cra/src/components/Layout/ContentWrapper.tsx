import styled, { css } from 'styled-components';
import { spacing } from 'weaveworks-ui-components/lib/theme/selectors';

export const pageDimensionsCss = css`
  padding: 0 ${spacing('xl')};
  max-width: 1400px;
`;

export const PageWrapper = styled.div`
  ${pageDimensionsCss}
  margin: ${spacing('large')} auto;
`;
