import { Button } from '@weaveworks/weave-gitops';
import {
  FlexCenter,
  Header4,
  LinkTag,
  OnBoardingMessageWrapper,
  TextWrapper,
} from '../CanaryStyles';

const OnboardingMessage = () => {
  return (
    <OnBoardingMessageWrapper>
      <Header4>Progressive Delivery</Header4>
      <TextWrapper>
        None of the clusters you have connected in Weave GitOps have the
        requirements installed for Progressive Delivery.
      </TextWrapper>
      <TextWrapper>
        To get started with this feature, follow the guide to install Flagger on
        your cluster(s).
      </TextWrapper>
      <Header4>Why Flagger?</Header4>
      <TextWrapper>
        Flagger was designed to give developers confidence in automating
        production releases with progressive delivery techniques. Flagger can
        run automated application analysis, testing, promotion, and rollback for
        deployment strategies such as Canary, A/B Testing, and Blue/Green.
      </TextWrapper>
      <FlexCenter>
        <Button id="navigate-to-flagger">
          <LinkTag href="https://flagger.app/" target="_blank" rel="noreferrer">
            FLAGGER DOCS
          </LinkTag>
        </Button>
      </FlexCenter>
    </OnBoardingMessageWrapper>
  );
};

export default OnboardingMessage;
