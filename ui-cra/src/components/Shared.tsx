import {
  DialogTitle,
  Tooltip as Mtooltip,
  TooltipProps,
  Typography,
} from '@material-ui/core';
import { Flex, Link } from '@weaveworks/weave-gitops';
import { FC } from 'react';
import styled from 'styled-components';
import CloseIconButton from '../assets/img/close-icon-button';

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
  width: 100%;
  div[class*='FilterDialog'] {
    .Mui-checked {
      color: ${({ theme }) => theme.colors.primary};
    }
    .MuiFormControl-root {
      min-width: 0px;
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

export const LinkTag = styled(Link)`
  color: ${({ theme }) => theme.colors.primary};
`;

type DialogTitleProps = {
  title: string;
  onFinish?: () => void;
};
export const MuiDialogTitle = ({ title, onFinish }: DialogTitleProps) => {
  return (
    <DialogTitle disableTypography>
      <Flex wide align between>
        <Typography variant="h5">{title}</Typography>
        {onFinish ? <CloseIconButton onClick={() => onFinish()} /> : null}
      </Flex>
    </DialogTitle>
  );
};
