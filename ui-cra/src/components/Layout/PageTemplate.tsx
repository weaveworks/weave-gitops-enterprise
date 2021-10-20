import React, { FC, useEffect } from 'react';
import { useDocumentTitle } from '../../utils/hooks';
import { PageWrapper } from './ContentWrapper';

interface Props {
  documentTitle?: string | null;
}

export const PageTemplate: FC<Props> = ({ children, documentTitle }) => {
  useDocumentTitle(documentTitle);

  useEffect(() => window.scrollTo(0, 0), []);

  return <PageWrapper>{children}</PageWrapper>;
};
