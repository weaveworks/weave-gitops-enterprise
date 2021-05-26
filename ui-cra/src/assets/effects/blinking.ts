import { css, keyframes } from 'styled-components';

const blink = keyframes`
  50% {
    opacity: 1;
  }
  75% {
    opacity: 0.4;
  }
  100% {
    opacity: 1;
  }
`;

export const blinking = css`
  animation: ${blink} 2.5s infinite;
`;
