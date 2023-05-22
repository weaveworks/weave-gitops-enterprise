import { Button, LoadingPage } from '@weaveworks/weave-gitops';
import { useCallback, useState } from 'react';
import styled from 'styled-components';
import useNotifications from '../../../contexts/Notifications';
import { PolicyConfigPRPreview } from '../../../types/custom';
import { renderKustomization } from '../../Applications/utils';
import Preview from '../../Templates/Form/Partials/Preview';

const PreviewPRSection = styled.div`
  display: flex;
  justify-content: flex-end;
  padding: ${props => props.theme.spacing.small};
`;

export const PreviewPRModal = ({ formData, getClusterAutomations }: any) => {
  const [openPreview, setOpenPreview] = useState(false);
  const [previewLoading, setPreviewLoading] = useState<boolean>(false);
  const [PRPreview, setPRPreview] = useState<PolicyConfigPRPreview | null>(
    null,
  );
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
    getClusterAutomations,
    setOpenPreview,
    setPRPreview,
    setPreviewLoading,
    setNotifications,
  ]);

  return (
    <PreviewPRSection>
      {previewLoading ? (
        <LoadingPage className="preview-loading" />
      ) : (
        <div className="preview-cta">
          <Button onClick={() => handlePRPreview()}>PREVIEW PR</Button>
        </div>
      )}
      {openPreview && PRPreview ? (
        <Preview
          context="policyconfig"
          openPreview={openPreview}
          setOpenPreview={setOpenPreview}
          PRPreview={PRPreview}
          sourceType={formData.source_type}
        />
      ) : null}
    </PreviewPRSection>
  );
};
