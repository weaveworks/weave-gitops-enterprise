import {
    Button,
    Flex,
    MessageBox,
    Text
} from '@weaveworks/weave-gitops';
import { NotificationsWrapper } from '../../Layout/NotificationsWrapper';
import { LinkTag } from '../../Shared';

const WarningMsg = () => {
  return (
    <NotificationsWrapper>
      <Flex center align>
        <MessageBox>
          <Flex column gap="20">
            <Text size="large" semiBold>
              Explorer Disabled
            </Text>
            <Text size="medium" capitalize>
              the explorer service is disabled and it's required to view the
              audit logs.
            </Text>
            <Flex wide align center>
              <LinkTag
                href="https://docs.gitops.weave.works/docs/explorer/configuration/"
                newTab
              >
                <Button id="navigate-to-imageautomation">
                  EXPLORER CONFIGRATION GUIDE
                </Button>
              </LinkTag>
            </Flex>
          </Flex>
        </MessageBox>
      </Flex>
    </NotificationsWrapper>
  );
};

export default WarningMsg;
