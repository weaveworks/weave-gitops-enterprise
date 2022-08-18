import { Button, Icon, IconType } from '@weaveworks/weave-gitops';
import React, { Dispatch, FC } from 'react';
import AppFields from '../../../../Applications/Add/form/Partials/AppFields';

export const ApplicationsWrapper: FC<{
  formData: any;
  setFormData: Dispatch<React.SetStateAction<any>>;
}> = ({ formData, setFormData }) => {
  const handleAddApplication = () => {
    let newKustomizations = [...formData.kustomizations];
    newKustomizations.push({ name: '', namespace: '', path: '' });
    setFormData({ ...formData, kustomizations: newKustomizations });
  };

  return (
    <>
      <h2>Applications</h2>
      {formData.kustomizations.map((kustomization: any, index: number) => {
        return (
          <AppFields
            key={index}
            formData={formData}
            setFormData={setFormData}
            kustomization={kustomization}
            index={index}
          ></AppFields>
        );
      })}
      <Button
        id="add-application"
        startIcon={<Icon type={IconType.AddIcon} size="base" />}
        onClick={handleAddApplication}
      >
        ADD AN APPLICATION
      </Button>
    </>
  );
};
