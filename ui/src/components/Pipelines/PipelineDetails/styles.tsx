import { Flex } from '@weaveworks/weave-gitops';
import styled from 'styled-components';

export const EnvironmentCard = styled(Flex)<{ background?: string }>`
  border-radius: 8px;
  padding: ${props => props.theme.spacing.small};
  width: 268px;
  height: 140px;
  color: black;
`;
