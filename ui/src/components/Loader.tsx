import { LoadingPage } from '@weaveworks/weave-gitops';
import React, { FC } from 'react';
import styled from 'styled-components';

const LoaderWrapper = styled.div`
  padding: ${({ theme }) => theme.spacing.medium} 0;
`;

export const Loader: FC = () => {
  return (
    <LoaderWrapper>
      <LoadingPage />
    </LoaderWrapper>
  );
};
