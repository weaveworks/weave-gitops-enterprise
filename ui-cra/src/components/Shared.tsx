import { Tooltip as Mtooltip, TooltipProps } from '@material-ui/core';
import { FC } from 'react';
import styled from 'styled-components';

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
  max-width: calc(100vw - 300px);
  div[class*='FilterDialog'] {
    .Mui-checked {
      color: ${({ theme }) => theme.colors.primary};
    }
  }
  div[class*='SearchField__Expander'] {
    overflow: hidden;
  }
  div.expanded {
    overflow: unset;
  }
`;

export const Title = styled.h4`
  font-size: ${({ theme }) => theme.fontSizes.large};
  font-weight: 600;
  color: ${({ theme }) => theme.colors.neutral30};
  margin-bottom: ${({ theme }) => theme.spacing.small};
`;

export const Body = styled.p`
  font-size: ${({ theme }) => theme.fontSizes.medium};
  color: ${({ theme }) => theme.colors.neutral30};
  font-weight: 400;
`;
