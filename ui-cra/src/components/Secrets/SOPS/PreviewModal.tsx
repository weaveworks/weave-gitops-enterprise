import { CircularProgress } from '@material-ui/core';
import { Button } from '@weaveworks/weave-gitops';
import { useCallback, useState } from 'react';
import useNotifications from '../../../contexts/Notifications';
import { SecretPRPreview } from '../../../types/custom';
import {
  encryptSopsSecret,
  renderKustomization
} from '../../Applications/utils';
import Preview from '../../Templates/Form/Partials/Preview';
import { PreviewPRSection } from './styles';
import { getFormattedPayload, handleError, SOPS } from './utils';

export const PreviewModal = ({ formData }: { formData: SOPS }) => {
  const [openPreview, setOpenPreview] = useState(false);
  const [previewLoading, setPreviewLoading] = useState<boolean>(false);
  const [PRPreview, setPRPreview] = useState<SecretPRPreview | null>(null);
  const { setNotifications } = useNotifications();

  const handlePRPreview = useCallback(async () => {
    setPreviewLoading(true);
    try {
      const { encryptionPayload, cluster } = getFormattedPayload(formData);
      const encrypted = await encryptSopsSecret(encryptionPayload);
      const render = await renderKustomization({
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
      setOpenPreview(true);
      setPRPreview(render);
    } catch (err: any) {
      handleError(err, setNotifications);
    } finally {
      setPreviewLoading(false);
    }
  }, [formData, setNotifications]);

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
