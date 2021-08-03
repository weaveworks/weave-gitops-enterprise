import React, { FC } from 'react';
import { useDocumentTitle } from '../../utils/hooks';
import { FooterWrapper } from './Footer';
import { PageWrapper } from './ContentWrapper';
import useNotifications from './../../contexts/Notifications';

interface Props {
  documentTitle?: string | null;
}

export const PageTemplate: FC<Props> = ({ children, documentTitle }) => {
  useDocumentTitle(documentTitle);
  const { notification } = useNotifications();

  return (
    <>
      <PageWrapper>{children}</PageWrapper>
      {notification ? <FooterWrapper notification={notification} /> : null}
    </>
  );
};
