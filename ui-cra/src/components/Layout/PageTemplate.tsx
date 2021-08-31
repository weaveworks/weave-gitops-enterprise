import React, { FC } from 'react';
import { useDocumentTitle } from '../../utils/hooks';
import { PageWrapper } from './ContentWrapper';

interface Props {
  documentTitle?: string | null;
}

export const PageTemplate: FC<Props> = ({ children, documentTitle }) => {
  useDocumentTitle(documentTitle);

  return <PageWrapper>{children}</PageWrapper>;
};
