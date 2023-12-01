import { Button } from '@weaveworks/weave-gitops';
import { Dispatch, useCallback, useState } from 'react';
import {
  ClustersService,
  RenderAutomationResponse,
} from '../../../cluster-services/cluster_services.pb';
import { useAPI } from '../../../contexts/API';
import useNotifications from '../../../contexts/Notifications';
import { validateFormData } from '../../../utils/form';
import PreviewModal from '../../Templates/Form/Partials/PreviewModal';
import {
  ExternalSecret,
  SOPS,
  getESFormattedPayload,
  getFormattedPayload,
  handleError,
} from './utils';

export enum SecretType {
  SOPS,
  ES,
}

const getRender = async (
  api: typeof ClustersService,
  secretType: SecretType,
  formData: SOPS | ExternalSecret,
) => {
  if (secretType === SecretType.SOPS) {
    const { encryptionPayload, cluster } = getFormattedPayload(
      formData as SOPS,
    );
    const encrypted = await api.EncryptSopsSecret(encryptionPayload);
    return await api.RenderAutomation({
      clusterAutomations: [
        {
          cluster,
          isControlPlane: cluster.namespace ? true : false,
          sopsSecret: {
            ...encrypted.encryptedSecret,
          },
          filePath: encrypted.path,
        },
      ],
    });
  } else {
    const payload = getESFormattedPayload(formData as ExternalSecret);
    return await api.RenderAutomation({
      clusterAutomations: [payload],
    });
  }
};

export const Preview = ({
  secretType = SecretType.SOPS,
  formData,
  setFormError,
}: {
  secretType?: SecretType;
  formData: SOPS | ExternalSecret;
  setFormError: Dispatch<React.SetStateAction<string>>;
}) => {
  const [openPreview, setOpenPreview] = useState(false);
  const [previewLoading, setPreviewLoading] = useState<boolean>(false);
  const [prPreview, setPRPreview] = useState<RenderAutomationResponse | null>(
    null,
  );
  const { setNotifications } = useNotifications();
  const { clustersService } = useAPI();

  const handlePRPreview = useCallback(async () => {
    setPreviewLoading(true);
    try {
      const render = getRender(enterprise, secretType, formData);
      setOpenPreview(true);
      setPRPreview(await render);
    } catch (err: any) {
      handleError(err, setNotifications);
    } finally {
      setPreviewLoading(false);
    }
  }, [enterprise, formData, secretType, setNotifications]);

  return (
    <>
      <Button
        onClick={event =>
          validateFormData(event, handlePRPreview, setFormError)
        }
        disabled={previewLoading}
        loading={previewLoading}
      >
        PREVIEW PR
      </Button>
      {!previewLoading && openPreview && prPreview ? (
        <PreviewModal
          context={secretType === SecretType.ES ? 'secret' : 'sops'}
          openPreview={openPreview}
          setOpenPreview={setOpenPreview}
          prPreview={prPreview}
        />
      ) : null}
    </>
  );
};
