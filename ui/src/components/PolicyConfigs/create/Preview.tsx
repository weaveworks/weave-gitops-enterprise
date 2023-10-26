import { Button } from '@weaveworks/weave-gitops';
import { Dispatch, useCallback, useContext, useState } from 'react';
import {
  ClusterAutomation,
  RenderAutomationResponse,
} from '../../../cluster-services/cluster_services.pb';
import { EnterpriseClientContext } from '../../../contexts/EnterpriseClient';
import useNotifications from '../../../contexts/Notifications';
import { validateFormData } from '../../../utils/form';
import PreviewModal from '../../Templates/Form/Partials/PreviewModal';

export const Preview = ({
  formData,
  getClusterAutomations,
  setFormError,
}: {
  formData: any;
  getClusterAutomations: () => ClusterAutomation[];
  setFormError: Dispatch<React.SetStateAction<string>>;
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
      .RenderAutomation({ clusterAutomations: getClusterAutomations() })
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
    api,
    getClusterAutomations,
    setOpenPreview,
    setPRPreview,
    setPreviewLoading,
    setNotifications,
  ]);

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
      {openPreview && prPreview ? (
        <PreviewModal
          context="policyconfig"
          openPreview={openPreview}
          setOpenPreview={setOpenPreview}
          prPreview={prPreview}
          sourceType={formData.source_type}
        />
      ) : null}
    </>
  );
};
