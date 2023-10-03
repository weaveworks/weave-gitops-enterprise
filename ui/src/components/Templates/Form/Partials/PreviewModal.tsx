import { Button } from '@weaveworks/weave-gitops';
import { useCallback, useState } from 'react';
import useTemplates from '../../../../hooks/templates';
import { TemplateEnriched } from '../../../../types/custom';
import {
  ProfileValues,
  RenderTemplateResponse,
  Credential,
  Kustomization,
} from '../../../../cluster-services/cluster_services.pb';
import useNotifications from '../../../../contexts/Notifications';
import Preview from './Preview';

export const PreviewModal = ({
  template,
  formData,
  profiles,
  credentials,
  kustomizations,
}: {
  template: TemplateEnriched;
  formData: any;
  profiles: ProfileValues[];
  credentials: Credential | undefined;
  kustomizations: Kustomization[];
}) => {
  const [openPreview, setOpenPreview] = useState(false);
  const [previewLoading, setPreviewLoading] = useState<boolean>(false);
  const [prPreview, setPRPreview] = useState<RenderTemplateResponse | null>(
    null,
  );
  const { setNotifications } = useNotifications();
  const { renderTemplate } = useTemplates();

  const handlePRPreview = useCallback(() => {
    const { parameterValues } = formData;
    setPreviewLoading(true);
    return renderTemplate({
      templateName: template.name,
      templateNamespace: template.namespace,
      values: parameterValues,
      profiles,
      credentials,
      kustomizations,
      templateKind: template.templateKind,
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
  }, [
    formData,
    setOpenPreview,
    renderTemplate,
    template.name,
    template.namespace,
    template.templateKind,
    setNotifications,
    profiles,
    credentials,
    kustomizations,
  ]);

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
          openPreview={openPreview}
          setOpenPreview={setOpenPreview}
          prPreview={prPreview}
        />
      ) : null}
    </>
  );
};
