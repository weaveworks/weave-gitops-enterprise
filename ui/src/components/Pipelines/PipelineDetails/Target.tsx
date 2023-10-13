import styled from 'styled-components';
import { PipelineTargetStatus } from '../../../api/pipelines/types.pb';

type Props = {
  className?: string;
  target: PipelineTargetStatus;
};

function Target({ className, target }: Props) {
  return <div className={className}></div>;
}

export default styled(Target).attrs({ className: Target.name })``;
