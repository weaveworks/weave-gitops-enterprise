import { Button } from '@weaveworks/weave-gitops';
import { useCallback, useContext, useState } from 'react';
import { EnterpriseClientContext } from '../../../contexts/EnterpriseClient';
import PreviewModal from '../../Templates/Form/Partials/PreviewModal';
import { RenderAutomationResponse } from '../../../cluster-services/cluster_services.pb';
import useNotifications from '../../../contexts/Notifications';

export const Preview = ({
  clusterAutomations,
}: {
  clusterAutomations: any;
}) => {
  const [openPreview, setOpenPreview] = useState(false);
  const [previewLoading, setPreviewLoading] = useState<boolean>(false);
  const [prPreview, setPRPreview] = useState<RenderAutomationResponse | null>(
    null,
  );
  const { setNotifications } = useNotifications();
  const { api } = useContext(EnterpriseClientContext);

  const handlePRPreview = useCallback(() => {
    setPreviewLoading(true);
    return api
      .RenderAutomation({
        clusterAutomations,
      })
      .then(data => {
        setOpenPreview(true);
        setPRPreview(data);
      })
      .catch(err =>
        setNotifications([
          {
            message: { text: err.message },
            severity: 'error',
            display: 'bottom',
          },
        ]),
      )
      .finally(() => setPreviewLoading(false));
  }, [api, setOpenPreview, clusterAutomations, setNotifications]);

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
        <PreviewModal
          openPreview={openPreview}
          setOpenPreview={setOpenPreview}
          prPreview={prPreview}
        />
      ) : null}
    </>
  );
};
