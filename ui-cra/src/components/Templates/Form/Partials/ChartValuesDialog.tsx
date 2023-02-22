import {
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  TextareaAutosize,
} from '@material-ui/core';
import { makeStyles } from '@material-ui/core/styles';
import Typography from '@material-ui/core/Typography';
import { Alert } from '@material-ui/lab';
import {
  Button,
  Icon,
  IconType,
  theme as weaveTheme,
} from '@weaveworks/weave-gitops';
import { ChangeEvent, FC, useContext } from 'react';
import { useQuery, useQueryClient } from 'react-query';
import { CloseIconButton } from '../../../../assets/img/close-icon-button';
import {
  GetConfigResponse,
  ClusterNamespacedName,
  GetChartsJobResponse,
  GetValuesForChartResponse,
  RepositoryRef,
} from '../../../../cluster-services/cluster_services.pb';
import { EnterpriseClientContext } from '../../../../contexts/EnterpriseClient';

import { UpdatedProfile } from '../../../../types/custom';
import { DEFAULT_PROFILE_REPO } from '../../../../utils/config';
import { Loader } from '../../../Loader';

const xs = weaveTheme.spacing.xs;

const useStyles = makeStyles(() => ({
  textarea: {
    width: '100%',
    padding: xs,
    border: `1px solid ${weaveTheme.colors.neutral10}`,
  },
}));

const ChartValuesDialog: FC<{
  cluster?: ClusterNamespacedName;
  yaml: string;
  profile: UpdatedProfile;
  version: string;
  onChange: (event: ChangeEvent<HTMLTextAreaElement>) => void;
  onSave: () => void;
  onClose: () => void;
  helmRepo: RepositoryRef;
}> = ({
  profile,
  yaml,
  version,
  cluster,
  onChange,
  onSave,
  onClose,
  helmRepo,
}) => {
  const classes = useStyles();
  const { api } = useContext(EnterpriseClientContext);
  const queryClient = useQueryClient();

  const getConfigResp = useQuery<GetConfigResponse, Error>('config', () =>
    api.GetConfig({}),
  );

  const {
    isLoading: jobLoading,
    data: jobData,
    error: jobError,
  } = useQuery<GetValuesForChartResponse, Error>(
    `values-job-${profile.name}-${version}`,
    () =>
      api.GetValuesForChart({
        repository: {
          cluster: cluster || {
            name: getConfigResp?.data?.managementClusterName,
          },
          name: helmRepo.name || DEFAULT_PROFILE_REPO.name,
          namespace: helmRepo.namespace || DEFAULT_PROFILE_REPO.namespace,
        },
        name: profile.name,
        version,
      }),
    {
      enabled: !yaml && !!getConfigResp?.data?.managementClusterName,
      refetchOnWindowFocus: false,
    },
  );

  const { isLoading: valuesLoading, data: jobResult } = useQuery<
    GetChartsJobResponse,
    Error
  >(
    `values-job-${jobData?.jobId}`,
    () =>
      api.GetChartsJob({
        jobId: jobData?.jobId,
      }),
    {
      enabled: Boolean(jobData?.jobId),
      refetchInterval: res => (!res?.error && !res?.values ? 2000 : false),
    },
  );

  const error = jobError?.message || jobResult?.error;
  const isLoading =
    !yaml &&
    (jobLoading || valuesLoading || (!jobResult?.error && !jobResult?.values));

  const resetQueryOnClose = () => {
    const query = `values-job-${jobData?.jobId}`;
    queryClient.removeQueries({ queryKey: query, exact: true });
    onClose();
  };
  return (
    <>
      <Dialog
        open
        maxWidth="md"
        fullWidth
        scroll="paper"
        onClose={resetQueryOnClose}
      >
        {error && <Alert severity="error">{error}</Alert>}
        <DialogTitle disableTypography>
          <Typography variant="h5">
            {profile.name} (version:{version})
          </Typography>
          <CloseIconButton onClick={resetQueryOnClose} />
        </DialogTitle>
        <DialogContent>
          {isLoading ? (
            <Loader />
          ) : (
            <TextareaAutosize
              className={classes.textarea}
              defaultValue={yaml || jobResult?.values}
              onChange={onChange}
            />
          )}
        </DialogContent>
        <DialogActions>
          <Button
            id="discard-yaml"
            startIcon={<Icon type={IconType.ClearIcon} size="base" />}
            onClick={resetQueryOnClose}
            disabled={profile.required && profile.editable !== true}
          >
            DISCARD CHANGES
          </Button>
          <Button
            id="edit-yaml"
            startIcon={<Icon type={IconType.SaveAltIcon} size="base" />}
            onClick={onSave}
            disabled={profile.required && profile.editable !== true}
          >
            SAVE CHANGES
          </Button>
        </DialogActions>
      </Dialog>
    </>
  );
};

export default ChartValuesDialog;
