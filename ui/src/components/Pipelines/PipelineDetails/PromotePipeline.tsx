import React from 'react';
import ShowChartIcon from '@material-ui/icons/ShowChart';
import { Button, Flex, Link, Text } from '@weaveworks/weave-gitops';
import { ApprovePromotionRequest } from '../../../api/pipelines/pipelines.pb';

import { useApprove } from '../../../hooks/pipelines';

const PromotePipeline = ({
  req,
  promoteVersion,
}: {
  req: ApprovePromotionRequest;
  promoteVersion: string;
}) => {
  const { data, isLoading, refetch } = useApprove(req);

  return (
    <Flex column gap="4">
      <Button
        startIcon={<ShowChartIcon />}
        onClick={() => refetch()}
        disabled={isLoading || !promoteVersion}
        loading={isLoading}
      >
        Approve Promotion
      </Button>
      {/* <Text color="primary20">
        PR:
        {data ? (
          <Link href={data.pullRequestURL} newTab>
            {data.pullRequestURL}
          </Link>
        ) : (
          ' Waiting For Approval'
        )}
      </Text> */}
    </Flex>
  );
};

export default PromotePipeline;
