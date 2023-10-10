import ShowChartIcon from '@material-ui/icons/ShowChart';
import { Button, Flex } from '@weaveworks/weave-gitops';
import React from 'react';
import styled from 'styled-components';
import { ApprovePromotionRequest } from '../../../api/pipelines/pipelines.pb';

import { useApprove } from '../../../hooks/pipelines';

const PromotePipeline = ({
  className,
  req,
  promoteVersion,
}: {
  className?: string;
  req: ApprovePromotionRequest;
  promoteVersion: string;
}) => {
  const approve = useApprove();
  return (
    <Flex column gap="8" className={className}>
      <Button
        startIcon={<ShowChartIcon />}
        onClick={() => approve.mutateAsync(req)}
        disabled={approve.isLoading || !promoteVersion}
        loading={approve.isLoading}
      >
        Approve Promotion
      </Button>
      {/** Add PR link here when backend changes get made */}
    </Flex>
  );
};

export default styled(PromotePipeline)``;
