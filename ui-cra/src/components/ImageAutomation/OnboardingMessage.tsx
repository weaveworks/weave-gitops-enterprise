import {
  Button,
  Flex,
  MessageBox,
  Spacer,
  Text,
} from '@weaveworks/weave-gitops';
import { LinkTag } from '../Shared';

const OnboardingImageAutomationMessage = () => {
  return (
    <MessageBox>
      <Spacer padding="small" />
      <Text size="large" semiBold>
        Image Automation
      </Text>
      <Spacer padding="small" />
      <Text size="medium">
        None of the clusters you have connected in Weave GitOps have the
        requirements installed for Image Automation.
      </Text>
      <Spacer padding="small" />
      <Text size="medium">
        To get started with this feature, follow the Flux guide to install the
        Image Reflector and Image Automation controllers on your cluster(s).
      </Text>
      <Spacer padding="small" />
      <Flex wide align center>
        <LinkTag href="https://fluxcd.io/flux/guides/image-update/" newTab>
          <Button id="navigate-to-imageautomation">
            IMAGE AUTOMATION GUIDE
          </Button>
        </LinkTag>
      </Flex>
    </MessageBox>
  );
};

export default OnboardingImageAutomationMessage;
