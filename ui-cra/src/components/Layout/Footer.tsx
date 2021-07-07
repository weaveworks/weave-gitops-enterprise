import React, { FC, Ref, useEffect, useRef } from 'react';
import styled from 'styled-components';
import { contentCss } from './ContentWrapper';

const Footer = styled.div`
  ${contentCss}
  display: flex;
  flex-direction: column;
  align-items: center;
  background-color: #ffcccc;
`;

export const FooterWrapper: FC<{ error: string }> = error => {
  const errorRef: Ref<HTMLDivElement> = useRef(null);

  useEffect(() => {
    if (error && errorRef.current) {
      errorRef.current.scrollIntoView({ behavior: 'smooth', block: 'center' });
    }
  }, [error]);

  return <Footer ref={errorRef}>{error.error}</Footer>;
};
