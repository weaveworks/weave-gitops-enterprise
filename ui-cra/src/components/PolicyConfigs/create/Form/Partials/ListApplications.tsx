import { Checkbox, FormControlLabel, FormGroup } from '@material-ui/core';
import { useListAutomations } from '@weaveworks/weave-gitops';
import { Dispatch, useEffect, useState } from 'react';
import { PolicyConfigApplicationMatch } from '../../../../../cluster-services/cluster_services.pb';
import { usePolicyConfigStyle } from '../../../PolicyConfigStyles';

interface SelectSecretStoreProps {
  cluster: string;
  formError: string;
  formData: any;
  selectedApplytList: any[];
  setSelectedApplytList: Dispatch<React.SetStateAction<any>>;
  setFormData: Dispatch<React.SetStateAction<any>>;
}
export const ListApplications = ({
  cluster,
  formData,
  selectedApplytList,
  setSelectedApplytList,
  setFormData,
}: SelectSecretStoreProps) => {
  const classes = usePolicyConfigStyle();
  const { appMatch = [] } = formData;
  const [isChecked, setIsChecked] = useState<string[]>([]);
  const [selectedTargetList, setSelectedTargetList] = useState<any[]>([]);
  const { data: applicationsList, isLoading: isApplicationsListLoading } =
    useListAutomations('', { retry: false });

  const applications =
    applicationsList?.result
      ?.filter(app => app.clusterName === cluster)
      .sort((a, b) => a.obj.metadata.name - b.obj.metadata.name) || [];

  useEffect(() => {
    if (formData.appMatch) {
      setSelectedTargetList(formData.appMatch);
      setIsChecked(
        appMatch.map((i: PolicyConfigApplicationMatch) => `${i.name}${i.kind}`),
      );
    }
  }, [formData.appMatch, setSelectedApplytList, appMatch]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>, app: any) => {
    setSelectedApplytList([]);

    const { name, checked } = e.target;
    let selected = selectedTargetList;
    if (checked) {
      selected.push({
        kind: app.obj.kind,
        name: name,
        namespace: app.obj.metadata.namespace || '',
      });
      setIsChecked([...isChecked, `${name}${app.obj.kind}`]);
    } else {
      selected = selected.filter(item => item.name !== name);
      setIsChecked(
        isChecked.filter((i: string) => i === `${name}${app.obj.kind}`),
      );
    }

    setSelectedTargetList(selected);
    setSelectedApplytList(selected);
    setFormData({
      ...formData,
      appMatch: selected,
    });
  };
  return cluster ? (
    <div>
      {!isApplicationsListLoading ? (
        applications.length ? (
          <FormGroup>
            <ul className={classes.checkList}>
              {applications.map(app => (
                <li key={app.obj.metadata.name}>
                  <FormControlLabel
                    control={
                      <Checkbox
                        size="small"
                        checked={isChecked.includes(
                          app.obj.metadata.name + app.obj.kind,
                        )}
                        name={app.obj.metadata.name}
                        onChange={e => handleChange(e, app)}
                      />
                    }
                    label={
                      <>
                        <span>
                          {app.obj.metadata.namespace === ''
                            ? '*'
                            : app.obj.metadata.namespace}
                          /{app.obj.metadata.name}
                        </span>
                        <span
                          className={`${classes.targetItemKind} ${classes.capitlize}`}
                        >
                          {app.obj.kind}
                        </span>
                      </>
                    }
                  />
                </li>
              ))}
            </ul>
          </FormGroup>
        ) : (
          <span>No Applications found</span>
        )
      ) : (
        <span>Loading...</span>
      )}
    </div>
  ) : (
    <span>No cluster selected yet</span>
  );
};