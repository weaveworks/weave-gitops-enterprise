import styled, { css } from 'styled-components';
import { theme } from '@weaveworks/weave-gitops';
import { Button as _Button } from 'weaveworks-ui-components';
import { darken } from 'polished';

export const ListView = styled.div`
  background: ${theme.colors.white};
  border-radius: ${theme.borderRadius.soft};
  box-sizing: border-box;
  color: ${theme.colors.neutral40};
  box-shadow: ${theme.boxShadow.light}; ;
`;

export const ListViewHeader = styled.div`
  margin: 0 ${theme.spacing.base} ${theme.spacing.base}} 0;

  display: flex;
  align-items: center;

  font-size: ${theme.fontSizes.large};

  & > * {
    margin-right: 8px;
  }
  & > *:last-child {
    margin-right: 0;
  }

  a {
    color: ${darken(0.1, 'hsl(0, 0%, 10%)')};
    display: block;
    font-size: ${theme.fontSizes.normal};
  }

  a:hover {
    color: ${theme.colors.primary};
  }
`;

export const Button = styled(_Button)`
  color: ${theme.colors.neutral40};
  margin: ${theme.spacing.xxs};
  min-height: ${theme.spacing.large};
  padding: ${theme.spacing.xxs} ${theme.spacing.xs};
  white-space: nowrap;
`;

interface ListItemProps {
  disabled?: boolean;
}
export const ListItem = styled.div<ListItemProps>`
  min-height: 52px;
  padding: 0 10px;
  border-bottom: 1px solid ${theme.colors.neutral30};
  display: flex;
  align-items: center;
  white-space: nowrap;

  ${props =>
    props.disabled
      ? css`
          color: ${theme.colors.neutral30};
          background-color: ${theme.colors.neutral20};
        `
      : ''}
`;

export const FlexSpacer = styled.div`
  flex: 1;
`;

export const Message = styled.div`
  padding: ${theme.spacing.base};
`;

interface IconWrapperProps {
  invertColors?: boolean;
}

export const IconWrapper = styled.div<IconWrapperProps>`
  ${props =>
    props.invertColors
      ? css`
          background-color: ${theme.colors.primary};
          color: ${theme.colors.white};
        `
      : css`
          background-color: ${theme.colors.neutral20};
          color: ${theme.colors.primary};
        `}
  display: flex;
  justify-content: center;
  align-items: center;
  border-radius: ${theme.borderRadius.soft};
  margin-right: 10px;
  font-size: ${theme.fontSizes.large};
  width: 32px;
  height: 32px;
  box-sizing: border-box;
`;
