import {
  Dialog,
  DialogContent,
  DialogTitle,
  Typography,
} from '@material-ui/core';
import { Link } from '@weaveworks/weave-gitops';
import { FC } from 'react';
import styled from 'styled-components';
import CloseIconButton from '../../assets/img/close-icon-button';

const DialogWrapper = styled(Dialog)`
  div[class*='MuiDialog-paper'] {
    padding: ${({ theme }) => theme.spacing.medium};
    border-radius: ${({ theme }) => theme.spacing.xs};
  }
`;

interface Props {
  onFinish: () => void;
}
export const ConnectClusterDialog: FC<Props> = ({ onFinish }) => {
  return (
    <DialogWrapper
      id="connection-popup"
      fullWidth
      maxWidth="md"
      onClose={() => onFinish()}
      open
    >
      <DialogTitle disableTypography>
        <Typography variant="h5">Connect a cluster</Typography>
        {onFinish ? <CloseIconButton onClick={() => onFinish()} /> : null}
      </DialogTitle>
      <DialogContent>
        For instructions on how to connect and disconnect clusters, have a look
        at the&nbsp;
        <Link
          href=" https://docs.gitops.weave.works/docs/cluster-management/managing-existing-clusters/#how-to-connect-a-cluster"
          newTab
        >
          documentation.
        </Link>
      </DialogContent>
    </DialogWrapper>
  );
};
