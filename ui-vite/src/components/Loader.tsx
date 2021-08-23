import React, { FC } from 'react';
import Skeleton from '@material-ui/lab/Skeleton';
import styled from 'styled-components';
import theme from 'weaveworks-ui-components/lib/theme';

const LoaderWrapper = styled.div`
  display: flex;
  justify-content: center;
  padding: ${theme.spacing.medium} 0;
`;

export const Loader: FC = () => {
  return (
    <LoaderWrapper>
      <Skeleton variant="circle" width={40} height={40} />
    </LoaderWrapper>
  );
};
