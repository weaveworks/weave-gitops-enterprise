import { Button, Flex } from '@weaveworks/weave-gitops';
import { Routes } from '../../../utils/nav';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import { SectionHeader } from '../../Layout/SectionHeader';
import {
  Header4,
  LinkTag,
  OnBoardingMessageWrapper,
  TextWrapper,
} from '../CanaryStyles';

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
          <Header4>Why Flagger?</Header4>
          <TextWrapper>
            Flagger was designed to give developers confidence in automating
            production releases with progressive delivery techniques. Flagger
            can run automated application analysis, testing, promotion, and
            rollback for deployment strategies such as Canary, A/B Testing, and
            Blue/Green.
          </TextWrapper>
          <Flex align center>
            <LinkTag
              href="https://docs.gitops.weave.works/docs/next/guides/delivery/"
              newTab
            >
              <Button id="navigate-to-flagger"> FLAGGER GUIDE</Button>
            </LinkTag>
          </Flex>
        </OnBoardingMessageWrapper>
      </ContentWrapper>
    </>
  );
};

export default OnboardingMessage;
