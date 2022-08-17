import React, { Dispatch, FC } from 'react';
import AppFields from '../../../../Applications/Add/form/Partials/AppFields';

export const ApplicationsWrapper: FC<{
  formData: any;
  setFormData: Dispatch<React.SetStateAction<any>>;
}> = ({ formData, setFormData }) => {
  return (
    <>
      <h2>Applications</h2>
      {formData.kustomizations.map((kustomization: any, index: number) => {
        return (
          <AppFields
            formData={kustomization}
            setFormData={setFormData}
          ></AppFields>
        );
      })}
    </>
  );
};
