import { Flex } from '@weaveworks/weave-gitops';
import styled from 'styled-components';

//blue, pink, then yellow
export const envColors = ['#E5F7FD', '#F6F4FF', '#FFFEF4'];

export const cardHeight = '140px';

export const EnvironmentCard = styled(Flex)<{ background?: number }>`
  border-radius: 8px;
  padding: ${props => props.theme.spacing.small};
  width: 268px;
  height: ${cardHeight};
  color: black;
  background: ${props =>
    //use remainder operator to alternate between items in envColors array
    envColors[props.background % envColors.length] || 'white'};
  box-shadow: 0px 4px 4px 0px rgba(0, 0, 0, 0.25);
  overflow: hidden;
`;
