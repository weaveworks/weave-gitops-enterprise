import { Button, LoadingPage } from '@weaveworks/weave-gitops';
import { useCallback, useState } from 'react';
import useNotifications from '../../../contexts/Notifications';
import { SecretPRPreview } from '../../../types/custom';
import Preview from '../../Templates/Form/Partials/Preview';

import { renderKustomization } from '../../Applications/utils';

export const PrviewPRModal = ({ formData, getClusterAutomations }: any) => {
  const [openPreview, setOpenPreview] = useState(false);
  const [previewLoading, setPreviewLoading] = useState<boolean>(false);
  const [PRPreview, setPRPreview] = useState<SecretPRPreview | null>(null);
  const { setNotifications } = useNotifications();

  const handlePRPreview = useCallback(() => {
    setPreviewLoading(true);
    return renderKustomization({ clusterAutomations: getClusterAutomations() })
      .then(data => {
        setOpenPreview(true);
        setPRPreview(data);
      })
      .catch(err => {
        setNotifications([
          {
            message: { text: err.message },
            severity: 'error',
            display: 'bottom',
          },
        ]);
      })
      .finally(() => setPreviewLoading(false));
  }, [
    formData,
    getClusterAutomations,
    setOpenPreview,
    setPRPreview,
    setPreviewLoading,
    setNotifications,
  ]);

  return (
    <div className="previewPRSection">
      {previewLoading ? (
        <LoadingPage className="preview-loading" />
      ) : (
        <div className="preview-cta">
          <Button onClick={() => handlePRPreview()}>PREVIEW PR</Button>
        </div>
      )}
      {openPreview && PRPreview ? (
        <Preview
          context="secret"
          openPreview={openPreview}
          setOpenPreview={setOpenPreview}
          PRPreview={PRPreview}
          sourceType={formData.source_type}
        />
      ) : null}
    </div>
  );
};
