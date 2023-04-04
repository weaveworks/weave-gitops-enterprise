import {
  Button,
  Flex,
  MessageBox,
  Spacer,
  Text,
} from '@weaveworks/weave-gitops';
import { Routes } from '../../../utils/nav';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import { SectionHeader } from '../../Layout/SectionHeader';
import { LinkTag } from '../../Shared';

const OnboardingMessage = () => {
  return (
    <>
      <SectionHeader
        className="count-header"
        path={[
          {
            label: 'Applications',
            url: Routes.Applications,
          },
          { label: 'Delivery' },
        ]}
      />
      <ContentWrapper>
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
              rollback for deployment strategies such as Canary, A/B Testing,
              and Blue/Green.
            </Text>
            <Spacer padding="small" />
            <Flex wide align center>
              <LinkTag
                href="https://docs.gitops.weave.works/docs/next/guides/delivery/"
                newTab
              >
                <Button id="navigate-to-flagger"> FLAGGER GUIDE</Button>
              </LinkTag>
            </Flex>
          </MessageBox>
        </Flex>
      </ContentWrapper>
    </>
  );
};

export default OnboardingMessage;
