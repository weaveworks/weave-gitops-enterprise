import React, { FC } from 'react';
import { useDocumentTitle } from '../../utils/hooks';
import { FooterWrapper } from './Footer';
import { PageWrapper } from './ContentWrapper';

interface Props {
  documentTitle?: string | null;
  error?: string | null;
}

export const PageTemplate: FC<Props> = ({ children, documentTitle, error }) => {
  useDocumentTitle(documentTitle);

  return (
    <>
      <PageWrapper>{children}</PageWrapper>
      {error ? <FooterWrapper error={error} /> : null}
    </>
  );
};
