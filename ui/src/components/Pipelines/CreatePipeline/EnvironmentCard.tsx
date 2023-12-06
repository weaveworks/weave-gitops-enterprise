import { Flex, Icon, IconType, Text } from '@weaveworks/weave-gitops';
import styled from 'styled-components';

const EnvironmentCardWrapper = styled(Flex)`
  background: ${props => props.theme.colors.primaryLight05};
  padding: 24px 12px;
  border-radius: 8px;
  box-shadow: 0px 4px 4px 0px #00000040;
`;

const TargetsWrapper = styled(Flex)`
  padding-top: 10px;
  margin-top: 6px;
  border-top: 1px solid ${props => props.theme.colors.neutral20};
`;

const EnvironmentCard = () => {
  return (
    <EnvironmentCardWrapper column gap="4">
      <Flex between wide>
        <Text semiBold>Dev</Text>
        <Icon type={IconType.SettingsIcon} size="medium" color="primary10" />
      </Flex>
      <Flex gap="4">
        <Text semiBold>Stategy:</Text>
        <Text>Pull request</Text>
      </Flex>
      <TargetsWrapper between wide>
        <Flex gap="4">
          <Text semiBold>Targets:</Text>
          <Text>02</Text>
        </Flex>
        <Icon type={IconType.KeyboardArrowRightIcon} size="medium" />
      </TargetsWrapper>
    </EnvironmentCardWrapper>
  );
};

export default EnvironmentCard;
