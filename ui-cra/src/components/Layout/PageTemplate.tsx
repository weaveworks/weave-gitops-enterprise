import { FC, useEffect } from 'react';
import { useDocumentTitle } from '../../utils/hooks';
import { Breadcrumb } from '../Breadcrumbs';
import { PageWrapper } from './ContentWrapper';
import { SectionHeader } from './SectionHeader';

interface Props {
  documentTitle?: string | null;
  path?: Breadcrumb[];
}

export const PageTemplate: FC<Props> = ({ children, documentTitle, path }) => {
  useDocumentTitle(documentTitle);

  useEffect(() => {
    if (process.env.NODE_ENV !== 'test') {
      window.scrollTo(0, 0);
    }
  }, []);

  return (
    <PageWrapper>
      {path?.length && <SectionHeader path={path} className='count-header'/>}
      {children}
    </PageWrapper>
  );
};
