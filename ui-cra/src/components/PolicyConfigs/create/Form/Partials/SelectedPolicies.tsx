import {
  Card,
  CardContent,
  FormControl,
  FormControlLabel,
  FormLabel,
  Radio,
  RadioGroup,
  TextField,
} from '@material-ui/core';
import { RemoveCircleOutline } from '@material-ui/icons';
import { Autocomplete } from '@material-ui/lab';
import { Icon, IconType, Text } from '@weaveworks/weave-gitops';
import {
  PolicyObj,
  PolicyParam,
} from '@weaveworks/weave-gitops/ui/lib/api/core/core.pb';
import { Dispatch, useEffect, useMemo, useState } from 'react';
import { ReactComponent as ErrorIcon } from '../../../../../assets/img/error.svg';
import { useListPolicies } from '../../../../../contexts/PolicyViolations';
import { Input } from '../../../../../utils/form';
import {
  PolicyDetailsCardWrapper,
  SectionTitle,
  usePolicyConfigStyle,
} from '../../../PolicyConfigStyles';

interface SelectSecretStoreProps {
  cluster: string;
  formData: any;
  setFormData: Dispatch<React.SetStateAction<any>>;
  formError: string;
}

export const SelectedPolicies = ({
  cluster,
  formData,
  setFormData,
  formError,
}: SelectSecretStoreProps) => {
  const classes = usePolicyConfigStyle();
  const [selectedPolicies, setSelectedPolicies] = useState<PolicyObj[]>([]);

  const { data } = useListPolicies({});

  const policiesList = useMemo(
    () =>
      data?.policies?.filter(p =>
        formData.isControlPlane
          ? p.clusterName === cluster
          : p.clusterName === `${formData.clusterNamespace}/${cluster}`,
      ) || [],
    [data?.policies, cluster, formData],
  );

  useEffect(() => {
    if (
      formData.policies &&
      data?.policies?.length &&
      selectedPolicies.length === 0
    ) {
      const selected: PolicyObj[] = policiesList.filter((p: PolicyObj) =>
        Object.keys(formData.policies).includes(p.id!),
      );
      setSelectedPolicies(selected);
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
    const areSameValues =
      type === 'array'
        ? JSON.stringify(
            value.split(/[\s,]+/).filter((i: string) => i !== ''),
          ) === JSON.stringify(defaultValue?.split(/[\s,]+/))
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
      formData.policies = {
        ...formData.policies,
        [id as string]: {
          parameters: {
            ...formData.policies[id]?.parameters,
            [name as string]:
              type === 'array'
                ? value.split(/[\s,]+/).filter((i: string) => i !== ' ')
                : value,
          },
        },
      };
    }

    setFormData({
      ...formData,
      policies: formData.policies,
    });
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

  const policiesInput = () => (
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
          <Text uppercase color="neutral30" style={{ marginBottom: '12px' }}>
            Select the policies to include in this policy config
          </Text>
          <TextField
            {...params}
            variant="outlined"
            name="policies"
            disabled={cluster === undefined}
            style={{ border: 'none !important' }}
            InputProps={{
              ...params.InputProps,
              endAdornment: (
                <Icon type={IconType.SearchIcon} size="medium" color="black" />
              ),
            }}
          />
        </>
      )}
    />
  );

  const getParameterField = (param: PolicyParam, id: string) => {
    const { type, name } = param;
    switch (type) {
      case 'boolean':
        return (
          <FormControl>
            <FormLabel id="demo-row-radio-buttons-group-label">
              {name}
            </FormLabel>
            <RadioGroup
              row
              aria-labelledby="demo-row-radio-buttons-group-label"
              name="row-radio-buttons-group"
              value={getValue(id!, param)}
              onChange={event => {
                handlePolicyParams(
                  event.target.value === 'true' ? true : false,
                  id!,
                  param,
                );
              }}
            >
              {formData.policies[id!]?.parameters[param.name!] && (
                <span className="modified">Modified</span>
              )}
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
        );
      default:
        return (
          <>
            {formData.policies[id!]?.parameters[name!] && (
              <span className="modified">Modified</span>
            )}
            <Input
              className="form-section"
              type={type === 'integer' ? 'number' : 'text'}
              name={name}
              label={name}
              defaultValue={getValue(id!, param)}
              onChange={event => {
                handlePolicyParams(event.target.value, id!, param);
              }}
            />
          </>
        );
    }
  };

  return (
    <>
      <div className="form-field policyField">
        <SectionTitle>
          Policies <span>({selectedPolicies?.length || 0})</span>
        </SectionTitle>
        {policiesInput()}
      </div>
      {formError === 'policies' &&
        JSON.stringify(formData.policies) === '{}' && (
          <div className={classes.errorSection}>
            <ErrorIcon />
            <span>Please add at least one policy with modified parameter</span>
          </div>
        )}

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
                      {getParameterField(param, policy.id!)}
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
