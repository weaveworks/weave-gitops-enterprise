import React, { FC } from 'react';
import styled from 'styled-components';
import { theme } from '@weaveworks/weave-gitops';
import { Tooltip as Mtooltip, TooltipProps } from '@material-ui/core';

export const Code = styled.div`
  display: flex;
  align-self: center;
  padding: 16px;
  background-color: ${theme.colors.white};
  font-family: ${theme.fontFamilies.monospace};
  border: 1px solid ${theme.colors.neutral20};
  border-radius: ${theme.borderRadius.soft};
  overflow: auto;
  font-size: ${theme.fontSizes.small};
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
