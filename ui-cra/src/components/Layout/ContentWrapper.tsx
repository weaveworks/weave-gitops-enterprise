import { FC } from 'react';
import styled, { css } from 'styled-components';
import {
  spacing,
  fontSize,
} from 'weaveworks-ui-components/lib/theme/selectors';
import theme from 'weaveworks-ui-components/lib/theme';
import useVersions from '../../contexts/Versions';
import { ReactComponent as WarningIcon } from '../../assets/img/warning-icon.svg';

export const Title = styled.div<{ extraPadding?: boolean }>`
  font-size: ${fontSize('large')};
  font-weight: 600;
  padding-bottom: ${props =>
    props.extraPadding ? spacing('xl') : spacing('medium')};
  color: ${theme.colors.gray600};
`;

export const pageDimensionsCss = css`
  width: 100%;
`;

export const PageWrapper = styled.div`
  ${pageDimensionsCss}
  margin: 0 auto;
`;

const small = spacing('small');
const medium = spacing('medium');
const large = spacing('large');

export const contentCss = css`
  margin: ${medium} ${small} 0 ${small};
  padding: ${large} ${medium} ${medium} ${large};
  background-color: white;
  border-radius: ${spacing('xs')};
`;

export const Content = styled.div<{ backgroundColor?: string }>`
  ${contentCss};
  background-color: ${props => props.backgroundColor};
`;

export const WGContent = styled.div`
  ${contentCss};
  padding: 0 ${medium} ${medium} 0;
`;

const EntitlementWrapper = styled.div`
  ${contentCss};
  background-color: #f3e9c9;
  padding: ${small} ${medium};
  display: flex;
`;

const WarningIconWrapper = styled(WarningIcon)`
  margin-right: ${small};
`;

const HelpLinkWrapper = styled.div`
  ${contentCss};
  padding: ${small} ${medium};
  color: ${({ theme }) => theme.colors.gray600};
  display: flex;
  justify-content: space-between;
  a {
    color: ${({ theme }) => theme.colors.blue600};
  }
`;

export const ContentWrapper: FC<{ type?: string; backgroundColor?: string }> =
  ({ children, type, backgroundColor }) => {
    const { versions, entitlement } = useVersions();
    return (
      <div style={{ display: 'flex', flexDirection: 'column', width: '100%' }}>
        {entitlement && (
          <EntitlementWrapper>
            <WarningIconWrapper />
            {entitlement}
          </EntitlementWrapper>
        )}
        {type === 'WG' ? (
          <WGContent>{children}</WGContent>
        ) : (
          <Content backgroundColor={backgroundColor}>{children}</Content>
        )}
        <HelpLinkWrapper>
          <div>
            Need help? Contact us at&nbsp;
            <a href="mailto:support@weave.works">support@weave.works</a>
          </div>
          <div>Version {versions?.capiServer}</div>
        </HelpLinkWrapper>
      </div>
    );
  };
