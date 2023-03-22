import {
    MenuItem
} from '@material-ui/core';
import { Dispatch, useState } from 'react';
import { PolicyConfigApplicationMatch } from '../../../../../cluster-services/cluster_services.pb';
import { Select } from '../../../../../utils/form';
import { usePolicyConfigStyle } from '../../../PolicyConfigStyles';
import { ListApplications } from './ListApplications';
import { ListWorkSpaces } from './ListWorkSpaces';

interface SelectSecretStoreProps {
  cluster: string;
  formError: string;
  formData: any;
  selectedWorkspacesList: string[];
  setSelectedWorkspacesList: Dispatch<React.SetStateAction<any>>;
  selectedAppsList: PolicyConfigApplicationMatch[];
  setSelectedAppsList: Dispatch<React.SetStateAction<any>>;
  handleFormData: (fieldName: string, value: any) => void;
  setFormData: Dispatch<React.SetStateAction<any>>;
}

export const SelectMatchType = ({
  cluster,
  formData,
  formError,
  selectedWorkspacesList,
  setSelectedWorkspacesList,
  selectedAppsList,
  setSelectedAppsList,
  handleFormData,
  setFormData,
}: SelectSecretStoreProps) => {
  const classes = usePolicyConfigStyle();
  const { matchType } = formData;

  const [matchTypeList, setMatchTypeList] = useState<string[]>([
    'workspaces',
    'apps',
  ]);

  const getCheckList = (matchType: string) => {
    switch (matchType) {
      case 'apps':
        return (
          <ListApplications
            cluster={cluster}
            formData={formData}
            formError={formError}
            selectedApplytList={selectedAppsList}
            setSelectedApplytList={setSelectedAppsList}
            setFormData={setFormData}
          />
        );
      case 'workspaces':
        return (
          <ListWorkSpaces
            cluster={cluster}
            formData={formData}
            formError={formError}
            selectedApplytList={selectedWorkspacesList}
            setSelectedApplytList={setSelectedWorkspacesList}
            setFormData={setFormData}
          />
        );
      default:
        <></>;
    }
  };

  // const HandleSelectTarget = (event: React.ChangeEvent<any>) => {
  //   const { value } = event.target;
  //   let currentAutomation = [...formData.clusterAutomations];
  //   currentAutomation[0] = {
  //     ...automation,
  //     matchType: value,
  //   };
  //   setFormData({
  //     ...formData,
  //     clusterAutomations: currentAutomation,
  //   });
  // };

  return (
    <>
      <div className="form-field">
        <label className={`${classes.sectionTitle}`}>Applied To</label>
        <Select
          className="form-section"
          name="matchType"
          placeholder="Select your target"
          label=""
          value={matchType || ''}
          onChange={e => handleFormData('matchType', e.target.value)}
          error={formError === 'matchType' && !matchType}
        >
          {matchTypeList?.map((option, index: number) => {
            return (
              <MenuItem
                key={index}
                value={option}
                className={classes.capitlize}
              >
                {option}
              </MenuItem>
            );
          })}
        </Select>
      </div>
      {getCheckList(matchType)}
    </>
  );
};
