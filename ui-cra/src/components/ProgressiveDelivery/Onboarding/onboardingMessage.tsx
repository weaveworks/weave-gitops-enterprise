import styled from 'styled-components';
import { Button, theme } from '@weaveworks/weave-gitops';
import { SectionHeader } from '../../Layout/SectionHeader';
import { ContentWrapper } from '../../Layout/ContentWrapper';

const OnBoardingMessageWrapper = styled.div`
  background: rgba(255, 255, 255, 0.85);
  box-shadow: 5px 10px 50px 3px rgb(0 0 0 / 10%);
  border-radius: 10px;
  padding: ${theme.spacing.large} ${theme.spacing.xxl};
  max-width: 560px;
  margin: auto;
`;

const Header4 = styled.div`
  font-size: ${theme.fontSizes.large};
  font-weight: 600;
  color: ${theme.colors.neutral30};
  margin-bottom: ${theme.spacing.small};
`;

const TextWrapper = styled.p`
  font-size: ${theme.fontSizes.normal};
  color: ${theme.colors.neutral30};
  font-weight: 400;
`;

const FlexCenter = styled.div`
  display: flex;
  lign-items: center;
  justify-content: center;
`;

const OnboardingMessage = () => {
  return (
    <div>
      <SectionHeader
        className="count-header"
        path={[
          { label: 'Applications', url: 'applications' },
          { label: 'Delivery', url: 'canaries' },
        ]}
      />
      <ContentWrapper>
        <OnBoardingMessageWrapper>
          <Header4>Progressive Delivery</Header4>
          <TextWrapper>
            None of the clusters you have connected in Weave GitOps have the
            requirements installed for Progressive Delivery.
          </TextWrapper>
          <TextWrapper>
            To get started with this feature, follow the guide to install
            Flagger on your cluster(s).
          </TextWrapper>
          <Header4>Progressive Delivery</Header4>
          <TextWrapper>
            Flagger was designed to give developers confidence in automating
            production releases with progressive delivery techniques. Flagger
            can run automated application analysis, testing, promotion, and
            rollback for deployment strategies such as Canary, A/B Testing, and
            Blue/Green.
          </TextWrapper>
          <FlexCenter>
            <Button id="navigate-to-flagger" onClick={() => {}}>
              <a href="https://flagger.app/" target="_blank" rel="noreferrer">
                FLAGGER DOCS
              </a>
            </Button>
          </FlexCenter>
        </OnBoardingMessageWrapper>
      </ContentWrapper>
    </div>
  );
};

export default OnboardingMessage;
