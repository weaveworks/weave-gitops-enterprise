import styled from 'styled-components';
import { PipelineTargetStatus } from '../../../api/pipelines/types.pb';
import { EnvironmentCard } from './styles';

type Props = {
  className?: string;
  targets: PipelineTargetStatus[];
};

const PromotionInfoContainer = styled(EnvironmentCard)`
  background: #f6f7f9;
  border: 1px solid black;
  border-style: dashed;
  color: ${props => props.theme.colors.black};
`;

function PromotionInfo({ className, targets }: Props) {
  return (
    <PromotionInfoContainer className={className}></PromotionInfoContainer>
  );
}

export default styled(PromotionInfo).attrs({ className: PromotionInfo.name })``;
