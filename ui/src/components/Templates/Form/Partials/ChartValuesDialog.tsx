import {
  Dialog,
  DialogActions,
  DialogContent,
  TextareaAutosize,
} from '@material-ui/core';
import { makeStyles } from '@material-ui/core/styles';
import { Alert } from '@material-ui/lab';
import { Button, Icon, IconType } from '@weaveworks/weave-gitops';
import { ChangeEvent, FC, useState } from 'react';
import { useQuery } from 'react-query';
import {
  ClusterNamespacedName,
  GetChartsJobResponse,
  GetConfigResponse,
  GetValuesForChartResponse,
  RepositoryRef,
} from '../../../../cluster-services/cluster_services.pb';
import { UpdatedProfile } from '../../../../types/custom';
import { DEFAULT_PROFILE_REPO } from '../../../../utils/config';
import { Loader } from '../../../Loader';
import { MuiDialogTitle } from '../../../Shared';
import { useAPI } from '../../../../contexts/API';

const useStyles = makeStyles(() => ({
  textarea: {
    width: '100%',
    padding: '4px',
    border: `1px solid #f5f5f5`,
  },
}));

const ChartValuesDialog: FC<{
  cluster?: ClusterNamespacedName;
  yaml: string;
  profile: UpdatedProfile;
  version: string;
  onSave: (value: string) => void;
  onClose: () => void;
  onDiscard: () => void;
  helmRepo: RepositoryRef;
}> = ({
  profile,
  yaml,
  version,
  cluster,
  onSave,
  onClose,
  helmRepo,
  onDiscard,
}) => {
  const classes = useStyles();
  const { enterprise } = useAPI();
  const [yamlPreview, setYamlPreview] = useState<string>(yaml);

  const getConfigResp = useQuery<GetConfigResponse, Error>('config', () =>
    enterprise.GetConfig({}),
  );

  const {
    isLoading: jobLoading,
    data: jobData,
    error: jobError,
  } = useQuery<GetValuesForChartResponse, Error>(
    `values-job-${profile.name}-${version}`,
    () =>
      enterprise.GetValuesForChart({
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
      refetchOnMount: false,
    },
  );

  const { isLoading: valuesLoading, data: jobResult } = useQuery<
    GetChartsJobResponse,
    Error
  >(
    `values-job-${jobData?.jobId}`,
    () =>
      enterprise.GetChartsJob({
        jobId: jobData?.jobId,
      }),
    {
      enabled: Boolean(jobData?.jobId),
      refetchInterval: res => (!res?.error && !res?.values ? 2000 : false),
      refetchOnMount: false,
    },
  );

  const error = jobError?.message || jobResult?.error;
  const isLoading =
    !yamlPreview &&
    (jobLoading || valuesLoading || (!jobResult?.error && !jobResult?.values));

  const handleYamlPreview = (event: ChangeEvent<HTMLTextAreaElement>) =>
    setYamlPreview(event.target.value);

  const handleModalClose = () => {
    setYamlPreview('');
    onClose();
  };

  const handleModalDiscard = () => {
    onSave('');
    onDiscard();
  };

  const handleModalSave = () => {
    onSave(yamlPreview);
  };

  return (
    <>
      <Dialog
        open
        maxWidth="md"
        fullWidth
        scroll="paper"
        onClose={handleModalClose}
      >
        {error && <Alert severity="error">{error}</Alert>}
        <MuiDialogTitle title={profile.name} onFinish={handleModalClose} />
        <DialogContent>
          {isLoading ? (
            <Loader />
          ) : (
            <TextareaAutosize
              className={classes.textarea}
              defaultValue={yamlPreview || jobResult?.values}
              onChange={handleYamlPreview}
            />
          )}
        </DialogContent>
        <DialogActions>
          <Button
            id="reset-yaml"
            startIcon={<Icon type={IconType.ClearIcon} size="base" />}
            onClick={handleModalDiscard}
            disabled={profile.required && profile.editable !== true}
          >
            RESET TO DEFAULT
          </Button>
          <Button
            id="edit-yaml"
            startIcon={<Icon type={IconType.SaveAltIcon} size="base" />}
            onClick={handleModalSave}
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
