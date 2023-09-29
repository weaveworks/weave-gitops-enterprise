import { MenuItem } from '@material-ui/core';
import { Flex, Text } from '@weaveworks/weave-gitops';
import { Dispatch } from 'react';
import { PolicyConfigApplicationMatch } from '../../../../../cluster-services/cluster_services.pb';
import { Select } from '../../../../../utils/form';
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
  const { matchType } = formData;
  const matchTypeList = ['workspaces', 'apps'];

  const getCheckList = (matchType: string) => {
    switch (matchType) {
      case 'apps':
        return (
          <ListApplications
            cluster={cluster}
            formData={formData}
            formError={formError}
            setSelectedAppsList={setSelectedAppsList}
            SelectedAppsList={selectedAppsList}
            setFormData={setFormData}
          />
        );
      case 'workspaces':
        return (
          <ListWorkSpaces
            cluster={cluster}
            formData={formData}
            formError={formError}
            selectedWorkspacesList={selectedWorkspacesList}
            setSelectedWorkspacesList={setSelectedWorkspacesList}
            setFormData={setFormData}
          />
        );
      default:
        <></>;
    }
  };

  return (
    <>
      <Flex column gap="16" className="form-field" wide>
        <Text capitalize semiBold size="large">
          Applied To
        </Text>
        <Select
          name="matchType"
          placeholder="Select your target"
          required
          label=""
          value={matchType || ''}
          onChange={e => handleFormData('matchType', e.target.value)}
          error={formError === 'matchType' && !matchType}
        >
          {matchTypeList?.map((option, index: number) => {
            return (
              <MenuItem key={index} value={option}>
                <Text size="base" capitalize>
                  {option}
                </Text>
              </MenuItem>
            );
          })}
        </Select>
      </Flex>
      {getCheckList(matchType)}
    </>
  );
};
