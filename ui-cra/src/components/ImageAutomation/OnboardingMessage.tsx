import { Button, Flex } from '@weaveworks/weave-gitops';
import {
  Header4,
  LinkTag,
  OnBoardingMessageWrapper,
  TextWrapper,
} from '../ProgressiveDelivery/CanaryStyles';

const OnboardingImageAutomationMessage = () => {
  return (
    <OnBoardingMessageWrapper>
      <Header4>Image Automation</Header4>
      <TextWrapper>
        None of the clusters you have connected in Weave GitOps have the
        requirements installed for Image Automation.
      </TextWrapper>
      <TextWrapper>
        To get started with this feature, follow the Flux guide to install the
        Image Reflector and Image Automation controllers on your cluster(s).
      </TextWrapper>
      <Flex align center>
        <LinkTag href="https://fluxcd.io/flux/guides/image-update/" newTab>
          <Button id="navigate-to-imageautomation">
            IMAGE AUTOMATION GUIDE
          </Button>
        </LinkTag>
      </Flex>
    </OnBoardingMessageWrapper>
  );
};

export default OnboardingImageAutomationMessage;
