import { CopyToClipboard, Flex } from '@weaveworks/weave-gitops';
import React from 'react';
import styled from 'styled-components';
interface Props {
  command: string;
}

const CommandText = styled.p`
  margin: 0;
  margin-right: ${props => props.theme.spacing.base};
  white-space: pre;
`;

const CommandCell = ({ command = '' }: Props) => {
  return (
    <Flex align>
      <CommandText>{command.replace(/--/g, '\\\n --')}</CommandText>
      <CopyToClipboard size="small" value={command} />
    </Flex>
  );
};

export default CommandCell;
