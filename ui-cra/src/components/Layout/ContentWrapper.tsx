import { FC } from 'react';
import styled, { css } from 'styled-components';
import {
  spacing,
  fontSize,
} from 'weaveworks-ui-components/lib/theme/selectors';
import theme from 'weaveworks-ui-components/lib/theme';
import useVersions from '../../contexts/Versions';
import { ReactComponent as WarningIcon } from '../../assets/img/warning-icon.svg';

export const Title = styled.div`
  font-size: ${fontSize('large')};
  font-weight: 600;
  padding-bottom: ${spacing('medium')};
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

export const Content = styled.div`
  ${contentCss};
`;

export const WGContent = styled.div`
  ${contentCss};
  padding: 0 ${medium} ${medium} 0;
`;

const HelpLinkWrapper = styled.div`
  padding-top: ${medium};
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  color: ${({ theme }) => theme.colors.gray600};
  white-space: nowrap;
  line-height: 1.5em;
  a {
    color: ${({ theme }) => theme.colors.blue600};
  }
`;

const EntitlementWrapper = styled.div`
  ${contentCss};
  background-color: #f3e9c9;
  padding: ${small} ${medium};
  display: flex;
  justifycontent: 'center';
`;

const WarningIconWrapper = styled(WarningIcon)`
  margin-right: ${small};
`;

export const ContentWrapper: FC<{ type?: string }> = ({ children, type }) => {
  const { versions, entitlement } = useVersions();
  return (
    <div style={{ display: 'flex', flexDirection: 'column' }}>
      {entitlement && (
        <EntitlementWrapper>
          <WarningIconWrapper />
          {entitlement}
        </EntitlementWrapper>
      )}
      {type === 'WG' ? (
        <WGContent>
          {children}
          <HelpLinkWrapper>
            <div>
              Need help? Contact us at&nbsp;
              <a href="mailto:support@weave.works">support@weave.works</a>
            </div>
            <div>Version {versions?.capiServer}</div>
          </HelpLinkWrapper>
        </WGContent>
      ) : (
        <Content>
          {children}
          <HelpLinkWrapper>
            <div>
              Need help? Contact us at{' '}
              <a href="mailto:support@weave.works">support@weave.works</a>
            </div>
            <div>Version {versions?.capiServer}</div>
          </HelpLinkWrapper>
        </Content>
      )}
    </div>
  );
};
