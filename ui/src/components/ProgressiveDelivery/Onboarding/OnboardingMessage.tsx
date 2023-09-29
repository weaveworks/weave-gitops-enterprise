import {
  Button,
  Flex,
  MessageBox,
  Spacer,
  Text,
} from '@weaveworks/weave-gitops';
import { NotificationsWrapper } from '../../Layout/NotificationsWrapper';
import { LinkTag } from '../../Shared';

const OnboardingMessage = () => {
  return (
    <NotificationsWrapper>
      <Flex center>
        <MessageBox>
          <Spacer padding="small" />
          <Text size="large" semiBold>
            Progressive Delivery
          </Text>
          <Spacer padding="small" />
          <Text size="medium">
            None of the clusters you have connected in Weave GitOps have the
            requirements installed for Progressive Delivery.
          </Text>
          <Spacer padding="xs" />
          <Text size="medium">
            To get started with this feature, follow the guide to install
            Flagger on your cluster(s).
          </Text>
          <Spacer padding="small" />
          <Text size="large" semiBold>
            Why Flagger?
          </Text>
          <Spacer padding="small" />
          <Text>
            Flagger was designed to give developers confidence in automating
            production releases with progressive delivery techniques. Flagger
            can run automated application analysis, testing, promotion, and
            rollback for deployment strategies such as Canary, A/B Testing, and
            Blue/Green.
          </Text>
          <Spacer padding="small" />
          <Flex wide align center>
            <LinkTag
              href="https://docs.gitops.weave.works/docs/guides/delivery/"
              newTab
            >
              <Button id="navigate-to-flagger"> FLAGGER GUIDE</Button>
            </LinkTag>
          </Flex>
        </MessageBox>
      </Flex>
    </NotificationsWrapper>
  );
};

export default OnboardingMessage;
