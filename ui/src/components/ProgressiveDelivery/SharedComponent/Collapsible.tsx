import { Collapse } from '@material-ui/core';
import { Flex, Icon, IconType, Text } from '@weaveworks/weave-gitops';
import React from 'react';
import styled from 'styled-components';

const CollapsibleDiv = styled.div`
  width: 100%;
  padding: 16px 4px;
  background: ${props => props.theme.colors.neutralGray};
  border-radius: 4px;
  cursor: pointer;
`;
const Collapsible = ({
  children,
  title,
}: {
  title?: string;
  children: any;
}) => {
  const [isOpen, setIsOpen] = React.useState(false);

  const toggle = () => setIsOpen(!isOpen);

  return (
    <div onClick={toggle} style={{ width: '100%' }}>
      <Flex column wide align>
        <CollapsibleDiv>
          <Flex wide align gap="16">
            <Icon
              type={
                isOpen
                  ? IconType.KeyboardArrowDownIcon
                  : IconType.KeyboardArrowRightIcon
              }
              size="medium"
              color="neutral40"
            />
            <Text color="neutral30">{title || 'More Information'}</Text>
          </Flex>
        </CollapsibleDiv>
        <Collapse in={isOpen} style={{ width: '100%' }}>
          {children}
        </Collapse>
      </Flex>
    </div>
  );
};

export default styled(Collapsible).attrs({})``;
