import { Button, Flex, LoadingPage } from '@weaveworks/weave-gitops';
import { useCallback, useContext, useState } from 'react';
import useNotifications from '../../../contexts/Notifications';
import Preview from '../../Templates/Form/Partials/Preview';
import {
  ExternalSecret,
  SOPS,
  getESFormattedPayload,
  getFormattedPayload,
  handleError,
} from './utils';
import { EnterpriseClientContext } from '../../../contexts/EnterpriseClient';
import {
  ClustersService,
  RenderAutomationResponse,
} from '../../../cluster-services/cluster_services.pb';

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

export const PreviewModal = ({
  secretType = SecretType.SOPS,
  formData,
}: {
  secretType?: SecretType;
  formData: SOPS | ExternalSecret;
}) => {
  const [openPreview, setOpenPreview] = useState(false);
  const [previewLoading, setPreviewLoading] = useState<boolean>(false);
  const [prPreview, setPRPreview] = useState<RenderAutomationResponse | null>(
    null,
  );
  const { setNotifications } = useNotifications();
  const { api } = useContext(EnterpriseClientContext);

  const handlePRPreview = useCallback(async () => {
    setPreviewLoading(true);
    try {
      const render = getRender(api, secretType, formData);
      setOpenPreview(true);
      setPRPreview(await render);
    } catch (err: any) {
      handleError(err, setNotifications);
    } finally {
      setPreviewLoading(false);
    }
  }, [api, formData, secretType, setNotifications]);

  return (
    <>
      <Button
        onClick={() => handlePRPreview()}
        disabled={previewLoading}
        loading={previewLoading}
      >
        PREVIEW PR
      </Button>
      {!previewLoading && openPreview && prPreview ? (
        <Preview
          context={secretType === SecretType.ES ? 'secret' : 'sops'}
          openPreview={openPreview}
          setOpenPreview={setOpenPreview}
          prPreview={prPreview}
        />
      ) : null}
    </>
  );
};
