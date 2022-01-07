import React, { FC } from 'react';
import styled from 'styled-components';
import { theme } from '@weaveworks/weave-gitops';
import { LoadingPage } from '@weaveworks/weave-gitops';

const LoaderWrapper = styled.div`
  padding: ${theme.spacing.medium} 0;
`;

export const Loader: FC = () => {
  return (
    <LoaderWrapper>
      <LoadingPage />
    </LoaderWrapper>
  );
};
