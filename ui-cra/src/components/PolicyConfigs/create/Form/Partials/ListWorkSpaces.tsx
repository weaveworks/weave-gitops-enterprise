import { Checkbox, FormControlLabel, FormGroup } from '@material-ui/core';
import { Dispatch, useEffect, useState } from 'react';
import { useListWorkspaces } from '../../../../../contexts/Workspaces';
import { usePolicyConfigStyle } from '../../../PolicyConfigStyles';

interface SelectSecretStoreProps {
  cluster: string;
  formError: string;
  formData: any;
  selectedApplytList: any[];
  setSelectedApplytList: Dispatch<React.SetStateAction<any>>;
  setFormData: Dispatch<React.SetStateAction<any>>;
}

export const ListWorkSpaces = ({
  cluster,
  formData,
  selectedApplytList,
  setSelectedApplytList,
  setFormData,
}: SelectSecretStoreProps) => {
  const classes = usePolicyConfigStyle();
  const [selectedTargetList, setSelectedTargetList] = useState<any[]>([]);

  const { data: workSpacesList, isLoading: isWorkSpacesListLoading } =
    useListWorkspaces({});
  const { wsMatch = [] } = formData;
  const workspaces =
    workSpacesList?.workspaces?.filter(
      workspace => workspace.clusterName === cluster,
    ) || [];

  useEffect(() => {
    if (formData.wsMatch) {
      setSelectedTargetList(formData.wsMatch);
    }
  }, [formData.wsMatch, setSelectedApplytList]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSelectedApplytList([]);
    const { name, checked } = e.target;
    if (checked) {
      setSelectedTargetList([...selectedTargetList, name]);
      formData.wsMatch = [...wsMatch, name];
    } else {
      setSelectedTargetList(selectedTargetList.filter(item => item !== name));
      formData.wsMatch = formData.wsMatch.filter(
        (item: string) => item !== name,
      );
    }
    setSelectedApplytList(formData.wsMatch);

    setFormData({
      ...formData,
      wsMatch: formData.wsMatch,
    });
  };

  return cluster ? (
    <div className="form-field">
      {!isWorkSpacesListLoading ? (
        workspaces.length ? (
          <FormGroup>
            <ul className={classes.checkList}>
              {workspaces.map(workspace => (
                <li
                  key={`${workspace.name}${workspace.clusterName}`}
                  className="workspaces"
                  // style={{ width: '33%', marginBottom: '0 !important' }}
                >
                  <FormControlLabel
                    key={workspace.name}
                    control={
                      <Checkbox
                        checked={selectedTargetList.includes(workspace.name)}
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
        )
      ) : (
        <span>Loading...</span>
      )}
    </div>
  ) : (
    <span>No cluster selected yet</span>
  );
};
