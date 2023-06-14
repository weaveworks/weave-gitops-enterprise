import { CircularProgress } from '@material-ui/core';
import { Button } from '@weaveworks/weave-gitops';
import { useCallback, useState } from 'react';
import useNotifications from '../../../contexts/Notifications';
import { SecretPRPreview } from '../../../types/custom';
import {
  encryptSopsSecret,
  renderKustomization,
} from '../../Applications/utils';
import Preview from '../../Templates/Form/Partials/Preview';
import { PreviewPRSection } from './styles';
import {
  ExternalSecret,
  getESFormattedPayload,
  getFormattedPayload,
  handleError,
  SOPS,
} from './utils';

export enum SecretType {
  SOPS,
  ES,
}
const getRender = async (
  secretType: SecretType,
  formData: SOPS | ExternalSecret,
) => {
  if (secretType === SecretType.SOPS) {
    const { encryptionPayload, cluster } = getFormattedPayload(
      formData as SOPS,
    );
    const encrypted = await encryptSopsSecret(encryptionPayload);
    return await renderKustomization({
      clusterAutomations: [
        {
          cluster,
          isControlPlane: cluster.namespace ? true : false,
          sops_secret: {
            ...encrypted.encryptedSecret,
          },
          file_path: encrypted.path,
        },
      ],
    });
  } else {
    const payload = getESFormattedPayload(formData as ExternalSecret);
    return await renderKustomization({
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
  const [PRPreview, setPRPreview] = useState<SecretPRPreview | null>(null);
  const { setNotifications } = useNotifications();

  const handlePRPreview = useCallback(async () => {
    setPreviewLoading(true);
    try {
      const render = getRender(secretType, formData);
      setOpenPreview(true);
      setPRPreview(await render);
    } catch (err: any) {
      handleError(err, setNotifications);
    } finally {
      setPreviewLoading(false);
    }
  }, [formData, secretType, setNotifications]);

  return (
    <PreviewPRSection>
      <div className="preview-cta">
        <Button onClick={() => handlePRPreview()} disabled={previewLoading}>
          PREVIEW PR
          {previewLoading && (
            <CircularProgress size={'1rem'} style={{ marginLeft: '4px' }} />
          )}
        </Button>
      </div>
      {openPreview && PRPreview ? (
        <Preview
          context="sops"
          openPreview={openPreview}
          setOpenPreview={setOpenPreview}
          PRPreview={PRPreview}
        />
      ) : null}
    </PreviewPRSection>
  );
};
