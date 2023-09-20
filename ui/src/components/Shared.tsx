import CloseIconButton from '../assets/img/close-icon-button';
import {
  DialogTitle,
  Tooltip as Mtooltip,
  TooltipProps,
  Typography,
} from '@material-ui/core';
import { Flex, Link } from '@weaveworks/weave-gitops';
import { FC } from 'react';
import ReactMarkdown from 'react-markdown';
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

export const Editor = styled(ReactMarkdown)`
  padding: ${props => props.theme.spacing.small};
  overflow: scroll;
  background: ${props => props.theme.colors.neutralGray};
  max-height: 300px;
  & a {
    color: ${props => props.theme.colors.primary};
  }

  & > *:first-child {
    margin-top: ${props => props.theme.spacing.none};
  }

  & > *:last-child {
    margin-bottom: ${props => props.theme.spacing.none};
  }
`;
