import React, { FC, Ref, useEffect, useRef, useState } from 'react';
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
  const [errorInView, setErrorInView] = useState<boolean | null>(null);
  const errorRef: Ref<HTMLDivElement> = useRef(null);

  useEffect(() => {
    if (error && errorRef.current && errorInView === false) {
      errorRef.current.scrollIntoView({ behavior: 'smooth', block: 'center' });
      return setErrorInView(true);
    }
  }, [error, errorInView]);

  return <Footer ref={errorRef}>{error.error}</Footer>;
};
