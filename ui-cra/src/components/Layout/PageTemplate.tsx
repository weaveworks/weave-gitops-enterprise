import React, { FC } from 'react';
import { useDocumentTitle } from '../../utils/hooks';
import { Footer } from './Footer';
import { PageWrapper } from './ContentWrapper';

interface Props {
  documentTitle?: string | null;
}

/**
 * PageTemplate provides the most common page layout including a NavBar etc.
 * If a page doesn't require the NavBar consider just using PageWrapper.
 */
export const PageTemplate: FC<Props> = ({ children, documentTitle }) => {
  useDocumentTitle(documentTitle);

  return (
    <>
      <PageWrapper>{children}</PageWrapper>
      <Footer />
    </>
  );
};
