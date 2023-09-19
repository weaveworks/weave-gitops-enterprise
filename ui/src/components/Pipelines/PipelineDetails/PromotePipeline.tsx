import React, { useCallback, useState } from 'react';
import ShowChartIcon from '@material-ui/icons/ShowChart';
import { Button, Flex, Link, Text } from '@weaveworks/weave-gitops';
import {
  ApprovePromotionRequest,
  Pipelines,
} from '../../../api/pipelines/pipelines.pb';
import { CircularProgress } from '@material-ui/core';
import { Alert, AlertTitle } from '@material-ui/lab';

const PromotePipeline = ({
  req,
  promoteVersion,
}: {
  req: ApprovePromotionRequest;
  promoteVersion: string;
}) => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [url, setUrl] = useState('');

  const approvePromotion = useCallback(() => {
    setLoading(true);
    setError('');
    Pipelines.ApprovePromotion(req)
      .then(res => {
        setUrl(res.pullRequestURL || '');
      })
      .catch(err => {
        setError(err?.message || 'Promoting fails');
      })
      .finally(() => {
        setLoading(false);
      });
  }, [req]);

  return (
    <>
      {error && !loading && (
        <Alert
          severity="error"
          style={{
            marginBottom: '8px',
          }}
        >
          <AlertTitle>Error promoting pipleline</AlertTitle>
          {error}
        </Alert>
      )}
      <Flex column gap="4">
        <Button
          startIcon={<ShowChartIcon />}
          onClick={() => approvePromotion()}
          disabled={loading || !promoteVersion}
          loading={loading}
        >
          Approve Promotion
        </Button>
        <Text color="primary20">
          PR:
          {url ? <Link href={url}>Pull Request</Link> : 'Waiting for approval'}
        </Text>
      </Flex>
    </>
  );
};

export default PromotePipeline;
