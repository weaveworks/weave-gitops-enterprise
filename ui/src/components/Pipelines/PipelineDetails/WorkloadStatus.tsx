import {
  Flex,
  Icon,
  IconType,
  KubeStatusIndicator,
} from '@weaveworks/weave-gitops';
import styled from 'styled-components';
import { WorkloadStatus as WorkloadStatusType } from '../../../api/pipelines/types.pb';

interface Props {
  className?: string;
  workload?: WorkloadStatusType;
}

function WorkloadStatus({ className, workload }: Props) {
  return (
    <div className={className}>
      <Flex align>
        <div style={{ marginRight: 4 }}>
          {workload?.suspended ? (
            <Icon
              size="small"
              type={IconType.SuspendedIcon}
              color="suspended"
            />
          ) : (
            <KubeStatusIndicator
              short
              conditions={workload?.conditions || []}
            />
          )}
        </div>
        <div>{workload?.name}</div>
      </Flex>
    </div>
  );
}

export default styled(WorkloadStatus).attrs({ className: 'WorkloadStatus' })`
  font-size: ${props => props.theme.fontSizes.large};

  /* Hack to hide the text; we only want the icon here */
  ${KubeStatusIndicator} span {
    display: none;
  }

  ${Icon} img {
    width: 16px;
  }
`;
