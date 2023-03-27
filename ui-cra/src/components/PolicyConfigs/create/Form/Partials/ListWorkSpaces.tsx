import { Checkbox, FormControlLabel, FormGroup } from '@material-ui/core';
import { Dispatch, useEffect } from 'react';
import { useListWorkspaces } from '../../../../../contexts/Workspaces';
import { usePolicyConfigStyle } from '../../../PolicyConfigStyles';

interface SelectSecretStoreProps {
  cluster: string;
  formError: string;
  formData: any;
  setSelectedWorkspacesList: Dispatch<React.SetStateAction<any>>;
  selectedWorkspacesList: any[];
  setFormData: Dispatch<React.SetStateAction<any>>;
}

export const ListWorkSpaces = ({
  cluster,
  formData,
  setSelectedWorkspacesList,
  setFormData,
  selectedWorkspacesList,
}: SelectSecretStoreProps) => {
  const { data: workSpacesList, isLoading: isWorkSpacesListLoading } =
    useListWorkspaces({});

  const classes = usePolicyConfigStyle();

  const workspaces =
    workSpacesList?.workspaces?.filter(
      workspace => workspace.clusterName === cluster,
    ) || [];

  useEffect(() => {
    if (formData.wsMatch) {
      setSelectedWorkspacesList(formData.wsMatch);
    }
  }, [formData.wsMatch, setSelectedWorkspacesList]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, checked } = e.target;
    let selected = selectedWorkspacesList;
    if (checked) {
      selected.push(name);
    } else {
      selected = selected.filter(item => item !== name);
    }
    setSelectedWorkspacesList(selected);
    setFormData({
      ...formData,
      wsMatch: selected,
    });
  };

  const getList = () => {
    switch (isWorkSpacesListLoading) {
      case false: {
        return workspaces.length ? (
          <FormGroup>
            <ul className={classes.checkList}>
              {workspaces.map(workspace => (
                <li
                  key={`${workspace.name}${workspace.clusterName}`}
                  className="workspaces"
                >
                  <FormControlLabel
                    key={workspace.name}
                    control={
                      <Checkbox
                        checked={selectedWorkspacesList.includes(
                          workspace.name,
                        )}
                        name={workspace.name}
                        onChange={e => handleChange(e)}
                      />
                    }
                    label={workspace.name}
                  />
                </li>
              ))}
            </ul>
          </FormGroup>
        ) : (
          <span>No Workspaces found</span>
        );
      }
      case true:
        return <span>Loading...</span>;
      default:
        return <></>;
    }
  };

  return cluster ? getList() : <span>No cluster selected yet</span>;
};