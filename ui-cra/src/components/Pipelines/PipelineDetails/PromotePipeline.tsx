import React, { useCallback, useState } from 'react';
import ShowChartIcon from '@material-ui/icons/ShowChart';
import { Button, Flex, Link } from '@weaveworks/weave-gitops';
import {
  ApprovePromotionRequest,
  Pipelines,
} from '../../../api/pipelines/pipelines.pb';
import { CircularProgress } from '@material-ui/core';

const PromotePipeline = ({
  req,
  promoteVersion,
}: {
  req: ApprovePromotionRequest;
  promoteVersion: string;
}) => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(false);
  const [url, setUrl] = useState('');

  const approvePromotion = useCallback(() => {
    setLoading(true);
    Pipelines.ApprovePromotion(req)
      .then(res => {
        setUrl(res.pullRequestURL || '');
      })
      .catch(() => {
        setError(true);
      })
      .finally(() => {
        setLoading(false);
      });
  }, [req]);

  if (error) {
    return <span>Something went wrong</span>;
  }
  return (
    <Flex align center >
      {!url ? (
        <Button
          startIcon={<ShowChartIcon />}
          onClick={() => approvePromotion()}
          disabled={loading}
        >
          Promote {promoteVersion}
          {loading && (
            <CircularProgress size={20} style={{ marginLeft: '8px' }} />
          )}
        </Button>
      ) : (
        <Link href={url}>Pull Request</Link>
      )}
    </Flex>
  );
};

export default PromotePipeline;
