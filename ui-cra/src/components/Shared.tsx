import React, { FC } from 'react';
import styled from 'styled-components';
import { Tooltip as Mtooltip, TooltipProps } from '@material-ui/core';

export const Code = styled.div`
  display: flex;
  align-self: center;
  padding: 16px;
  background-color: ${({ theme }) => theme.colors.white};
  font-family: ${({ theme }) => theme.fontFamilies.monospace};
  border: 1px solid ${({ theme }) => theme.colors.neutral20};
  border-radius: ${({ theme }) => theme.borderRadius.soft};
  overflow: auto;
  font-size: ${({ theme }) => theme.fontSizes.small};
`;

const TooltipStyle = styled.div`
  font-size: 14px;
`;

export const Tooltip: FC<TooltipProps & { disabled?: boolean }> = ({
  disabled,
  title,
  children,
  ...props
}) => {
  const styledTitle = <TooltipStyle>{title}</TooltipStyle>;
  return disabled ? (
    children
  ) : (
    <Mtooltip enterDelay={500} title={styledTitle} {...props}>
      {children}
    </Mtooltip>
  );
};

export const ColumnHeaderTooltip: FC<TooltipProps> = ({
  title,
  children,
  ...props
}) => (
  <Tooltip title={title} placement="top" {...props}>
    {children}
  </Tooltip>
);

export const TableWrapper = styled.div`
  margin-top: ${({ theme }) => theme.spacing.medium};
  div[class*='FilterDialog__SlideContainer'],
  div[class*='SearchField'] {
    overflow: hidden;
  }
  div[class*='FilterDialog'] {
    .Mui-checked {
      color: ${({ theme }) => theme.colors.primary};
    }
  }
  max-width: calc(100vw - 220px);
`;
