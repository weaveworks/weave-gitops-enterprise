import {
  Card,
  CardContent,
  FormControl,
  FormControlLabel,
  FormLabel,
  Radio,
  RadioGroup,
  TextField
} from '@material-ui/core';
import { RemoveCircleOutline } from '@material-ui/icons';
import SearchIcon from '@material-ui/icons/Search';
import { Autocomplete } from '@material-ui/lab';
import { Dispatch, useEffect, useMemo, useState } from 'react';
import {
  Policy,
  PolicyParam
} from '../../../../../cluster-services/cluster_services.pb';
import { useListListPolicies } from '../../../../../contexts/PolicyViolations';
import { Input } from '../../../../../utils/form';
import {
  PolicyDetailsCardWrapper,
  usePolicyConfigStyle
} from '../../../PolicyConfigStyles';

interface SelectSecretStoreProps {
  cluster: string;
  formError: string;
  formData: any;
  setFormData: Dispatch<React.SetStateAction<any>>;
}

export const SelectedPolicies = ({
  cluster,
  formError,
  formData,
  setFormData,
}: // automation,
SelectSecretStoreProps) => {
  const classes = usePolicyConfigStyle();
  // const { policies = {} } = formData;
  const [selectedPolicies, setSelectedPolicies] = useState<Policy[]>([]);

  const { data } = useListListPolicies({});

  const policiesList = useMemo(
    () => data?.policies?.filter(p => p.clusterName === cluster) || [],
    [data?.policies, cluster],
  );
  useEffect(() => {
    if (
      formData.policies &&
      data?.policies?.length &&
      selectedPolicies.length === 0
    ) {
      const selected: Policy[] = policiesList.filter((p: Policy) =>
        Object.keys(formData.policies).includes(p.id!),
      );
      setSelectedPolicies(selected);
      console.log('d');
    }
  }, [
    data?.policies,
    policiesList,
    formData.policies,
    selectedPolicies.length,
  ]);

  const handlePolicyParams = (val: any, id: string, param: PolicyParam) => {
    const { name, type } = param;
    const defaultValue =
      type === 'array' ? param.value?.value.join(', ') : param.value?.value;
    const value = type === 'integer' ? parseInt(val) || '0' : val;

    console.log(val, value, defaultValue);
    const areSameValues =
      type === 'array'
        ? JSON.stringify(value.split(',').filter((i: string) => i !== '')) ===
          JSON.stringify(defaultValue?.split(','))
        : value === defaultValue;

    if (
      areSameValues ||
      (value === '' && defaultValue === (null || undefined))
    ) {
      const policyConfigs = formData.policies;
      delete policyConfigs[id].parameters[name as string];
      if (Object.keys(policyConfigs[id]?.parameters).length === 0)
        delete policyConfigs[id];
    } else {
      console.log('else');
      formData.policies = {
        ...formData.policies,
        [id as string]: {
          parameters: {
            ...formData.policies[id]?.parameters,
            [name as string]:
              type === 'array'
                ? value.split(',').filter((i: string) => i !== '')
                : value,
          },
        },
      };
    }

    setFormData({
      ...formData,
      policies: formData.policies,
    });
    console.log(formData.policies, selectedPolicies);
  };
  const handleDeletePolicyParam = (id: string) => {
    const item = formData.policies || {};
    if (Object.keys(item).length !== 0) delete item[id];

    let updateSelected = selectedPolicies?.filter(p => p.id !== id);
    setSelectedPolicies(updateSelected);
  };
  const getValue = (id: string, param: PolicyParam) => {
    const isModified = formData.policies[id!]?.parameters[param.name!]
      ? true
      : false;
    const { type, name, value } = param;
    if (isModified) {
      switch (type) {
        case 'array':
          return formData.policies[id!].parameters[name!].join(', ');
        case 'integer':
          return parseInt(formData.policies[id!].parameters[name!]);
        default:
          return formData.policies[id!].parameters[name!].toString();
      }
    } else {
      switch (type) {
        case 'array':
          return value?.value.join(', ');
        case 'integer':
          return parseInt(value?.value);
        default:
          return value?.value.toString();
      }
    }
  };

  return (
    <>
      <div className="form-field policyField">
        <label className={classes.sectionTitle}>
          Policies <span>({selectedPolicies?.length || 0})</span>
        </label>
        <Autocomplete
          multiple
          className={classes.SelectPoliciesWithSearch}
          id="grouped-demo"
          value={selectedPolicies}
          options={policiesList?.sort((a, b) =>
            b.category!.localeCompare(a.category!),
          )}
          groupBy={option => option.category || ''}
          onChange={(e, policy) => setSelectedPolicies(policy)}
          noOptionsText="No Policies found on that cluster."
          getOptionLabel={option => option.name || ''}
          filterSelectedOptions
          renderInput={params => (
            <>
              <span className={classes.fieldNote}>
                Select the policies to include in this policy config
              </span>
              <TextField
                {...params}
                variant="outlined"
                name="policies"
                disabled={cluster === undefined}
                style={{ border: 'none !important' }}
                InputProps={{
                  ...params.InputProps,
                  endAdornment: (
                    <>
                      <SearchIcon />
                    </>
                  ),
                }}
              />
            </>
          )}
        />
      </div>
      <PolicyDetailsCardWrapper>
        {selectedPolicies?.map(policy => (
          <li key={policy.id}>
            <Card>
              <CardContent>
                <div className={`${classes.policyTitle} editPolicyCardHeader`}>
                  <span>{policy.name}</span>

                  <RemoveCircleOutline
                    onClick={() => handleDeletePolicyParam(policy.id!)}
                  />
                </div>
                <label className="cardLbl">Parameters</label>
                {policy?.parameters?.map(param => (
                  <div
                    className="parameterItem"
                    key={`${param.name}${policy.id}`}
                  >
                    <div className={`parameterItemValue ${classes.upperCase}`}>
                      {param.type === 'boolean' ? (
                        <FormControl>
                          <FormLabel id="demo-row-radio-buttons-group-label">
                            {param.name}
                          </FormLabel>
                          <RadioGroup
                            row
                            aria-labelledby="demo-row-radio-buttons-group-label"
                            name="row-radio-buttons-group"
                            value={getValue(policy.id!, param)}
                            onChange={event => {
                              handlePolicyParams(
                                event.target.value === 'true' ? true : false,
                                policy.id!,
                                param,
                              );
                            }}
                          >
                            {formData.policies[policy.id!]?.parameters[
                              param.name!
                            ] && <span className="modified">Modified</span>}
                            <FormControlLabel
                              value={'true'}
                              control={<Radio />}
                              label="True"
                            />
                            <FormControlLabel
                              value={'false'}
                              control={<Radio />}
                              label="False"
                            />
                          </RadioGroup>
                        </FormControl>
                      ) : (
                        <>
                          {formData.policies[policy.id!]?.parameters[
                            param.name!
                          ] && <span className="modified">Modified</span>}
                          <Input
                            className="form-section"
                            type={param.type === 'integer' ? 'number' : 'text'}
                            name={param.name}
                            label={param.name}
                            defaultValue={getValue(policy.id!, param)}
                            onChange={event => {
                              handlePolicyParams(
                                // param.type === 'integer'
                                //   ? parseInt(event.target.value) || '0'
                                //   :
                                event.target.value,
                                policy.id!,
                                param,
                              );
                            }}
                          />
                        </>
                      )}
                    </div>
                  </div>
                ))}
              </CardContent>
            </Card>
          </li>
        ))}
      </PolicyDetailsCardWrapper>
    </>
  );
};
