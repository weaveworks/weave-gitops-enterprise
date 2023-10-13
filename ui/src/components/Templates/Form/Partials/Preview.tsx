import { Button } from '@weaveworks/weave-gitops';
import { Dispatch, useCallback, useState } from 'react';
import {
  ProfileValues,
  RenderTemplateResponse,
  Credential,
  Kustomization,
} from '../../../../cluster-services/cluster_services.pb';
import useNotifications from '../../../../contexts/Notifications';
import useTemplates from '../../../../hooks/templates';
import { TemplateEnriched } from '../../../../types/custom';
import { validateFormData } from '../../../../utils/form';
import PreviewModal from './PreviewModal';

export const Preview = ({
  template,
  formData,
  profiles,
  credentials,
  kustomizations,
  setFormError,
}: {
  template: TemplateEnriched;
  formData: any;
  profiles: ProfileValues[];
  credentials: Credential | undefined;
  kustomizations: Kustomization[];
  setFormError: Dispatch<React.SetStateAction<string>>;
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

  console.log(prPreview);

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
          openPreview={openPreview}
          setOpenPreview={setOpenPreview}
          prPreview={prPreview}
        />
      ) : null}
    </>
  );
};
