import { FC } from 'react';
import styled, { css } from 'styled-components';
import { theme } from '@weaveworks/weave-gitops';
import useVersions from '../../contexts/Versions';
import { ReactComponent as WarningIcon } from '../../assets/img/warning-icon.svg';

const xs = theme.spacing.xs;
const small = theme.spacing.small;
const medium = theme.spacing.medium;
const large = theme.spacing.large;

export const Title = styled.h2`
  margin-top: 0px;
`;

export const pageDimensionsCss = css`
  width: 100%;
`;

export const PageWrapper = styled.div`
  ${pageDimensionsCss}
  margin: 0 auto;
`;

export const contentCss = css`
  margin: ${medium} ${small} 0 ${small};
  padding: ${large} ${medium} ${medium} ${large};
  background-color: ${theme.colors.white};
  border-radius: ${xs};
`;

export const Content = styled.div<{ backgroundColor?: string }>`
  ${contentCss};
  background-color: ${props => props.backgroundColor};
`;

export const WGContent = styled.div`
  margin: ${medium} ${small} 0 ${small};
  background-color: ${theme.colors.white};
  border-radius: ${xs};
  > div > div {
    border-radius: ${xs};
  }
`;

const EntitlementWrapper = styled.div`
  ${contentCss};
  background-color: ${theme.colors.feedbackLight};
  padding: ${small} ${medium};
  display: flex;
`;

const WarningIconWrapper = styled(WarningIcon)`
  margin-right: ${small};
`;

const HelpLinkWrapper = styled.div`
  ${contentCss};
  padding: ${small} ${medium};
  color: ${({ theme }) => theme.colors.neutral40};
  display: flex;
  justify-content: space-between;
  a {
    color: ${({ theme }) => theme.colors.primary};
  }
`;

export const ContentWrapper: FC<{
  type?: string;
  backgroundColor?: string;
}> = ({ children, type, backgroundColor }) => {
  const { versions, entitlement } = useVersions();
  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        width: '100%',
      }}
    >
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
