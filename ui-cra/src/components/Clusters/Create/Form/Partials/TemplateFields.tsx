import React, { FC, useMemo, useState, Dispatch, useEffect } from 'react';
import weaveTheme from 'weaveworks-ui-components/lib/theme';
import { Button } from 'weaveworks-ui-components';
import { makeStyles, createStyles } from '@material-ui/core/styles';
import {
  Template,
  TemplateObject,
  UpdatedProfile,
} from '../../../../../types/custom';
import { ObjectFieldTemplateProps } from '@rjsf/core';
import { JSONSchema7 } from 'json-schema';
import Form from '@rjsf/material-ui';
import * as Grouped from '../GroupedSchema';
import * as UiTemplate from '../UITemplate';
import FormSteps, { FormStep } from '../Steps';
import MultiSelectDropdown from '../../../../MultiSelectDropdown';
import ProfilesList from './ProfilesList';
import useProfiles from '../../../../../contexts/Profiles';

const base = weaveTheme.spacing.base;
const small = weaveTheme.spacing.small;

const useStyles = makeStyles(() =>
  createStyles({
    form: {
      paddingTop: base,
    },
    create: {
      paddingTop: small,
    },
    previewCTA: {
      display: 'flex',
      justifyContent: 'flex-end',
      paddingTop: small,
      paddingBottom: base,
    },
  }),
);

const TemplateFields: FC<{
  activeTemplate: Template | null;
  onPRPreview: () => void;
  activeStep: string | undefined;
  setActiveStep: Dispatch<React.SetStateAction<string | undefined>>;
  clickedStep: string;
  onProfilesUpdate: Dispatch<React.SetStateAction<UpdatedProfile[]>>;
  onFormDataUpdate: Dispatch<React.SetStateAction<any>>;
  formData: any;
}> = ({
  activeTemplate,
  onPRPreview,
  activeStep,
  setActiveStep,
  clickedStep,
  onProfilesUpdate,
  formData,
  onFormDataUpdate,
}) => {
  const classes = useStyles();
  const { updatedProfiles } = useProfiles();
  const [selectedProfiles, setSelectedProfiles] = useState<UpdatedProfile[]>(
    [],
  );

  const objectTitle = (object: TemplateObject, index: number) => {
    if (object.displayName && object.displayName !== '') {
      return `${index + 1}.${object.kind} (${object.displayName})`;
    }
    return `${index + 1}.${object.kind}`;
  };

  const required = useMemo(() => {
    return activeTemplate?.parameters?.map(param => param.name);
  }, [activeTemplate]);

  const parameters = useMemo(() => {
    // const p = [] as any;
    return (
      activeTemplate?.parameters?.map(param => {
        const { name, options } = param;
        if (options?.length !== 0) {
          return {
            [name]: {
              type: 'string',
              title: `${name}`,
              enum: options,
              // visible: true if it's the first param of the kind that appears
            },
          };
        } else {
          return {
            [name]: {
              type: 'string',
              title: `${name}`,
              // visible: true if it's the first param of the kind that appears
            },
          };
        }
      }) || []
    );

    // activeTemplate?.parameters?.map(param => {
    //   const { name, options } = param;

    //   if (options?.length !== 0) {
    //     return {
    //       [name]: {
    //         type: 'string',
    //         title: `${name}`,
    //         enum: options,
    //         // visible: true if it's the first param of the kind that appears
    //       },
    //     };
    //   } else {
    //     return {
    //       [name]: {
    //         type: 'string',
    //         title: `${name}`,
    //         // visible: true if it's the first param of the kind that appears
    //       },
    //     };
    //   }
    // });
    // return p;
  }, [activeTemplate]);

  const properties = useMemo(() => {
    return Object.assign({}, ...parameters);
  }, [parameters]);

  // console.log(properties);

  const schema: JSONSchema7 = useMemo(() => {
    return {
      type: 'object',
      properties,
      required,
    };
  }, [properties, required]);

  // Adapted from : https://codesandbox.io/s/0y7787xp0l?file=/src/index.js:1507-1521
  const sections = useMemo(() => {
    const groups =
      activeTemplate?.objects?.reduce(
        (accumulator, item, index) =>
          Object.assign(accumulator, {
            [objectTitle(item, index)]: item.parameters,
          }),
        {},
      ) || {};
    Object.assign(groups, { 'ui:template': 'box' });
    return [groups];
  }, [activeTemplate]);

  // console.log(sections);

  // accumulator: {1.Cluster: Array(1), 2.AWSCluster: Array(3), 3.KubeadmControlPlane: Array(3), 4.AWSMachineTemplate: Array(3), 5.MachineDeployment: Array(3), …}1.Cluster: ['CLUSTER_NAME']2.AWSCluster: (3) ['AWS_REGION', 'AWS_SSH_KEY_NAME', 'CLUSTER_NAME']3.KubeadmControlPlane: (3) ['CLUSTER_NAME', 'CONTROL_PLANE_MACHINE_COUNT', 'KUBERNETES_VERSION']4.AWSMachineTemplate: (3) ['AWS_CONTROL_PLANE_MACHINE_TYPE', 'AWS_SSH_KEY_NAME', 'CLUSTER_NAME']5.MachineDeployment: (3) ['CLUSTER_NAME', 'KUBERNETES_VERSION', 'WORKER_MACHINE_COUNT']6.AWSMachineTemplate: (3) ['AWS_NODE_MACHINE_TYPE', 'AWS_SSH_KEY_NAME', 'CLUSTER_NAME']7.KubeadmConfigTemplate: ['CLUSTER_NAME']ui:template: "box"[[Prototype]]: Object

  // 1.Cluster: Array(1)
  //     0: "CLUSTER_NAME"

  // item.parameters: (3) ['AWS_NODE_MACHINE_TYPE', 'AWS_SSH_KEY_NAME', 'CLUSTER_NAME']

  const uiSchema = useMemo(() => {
    return {
      'ui:groups': sections,
      'ui:template': (props: ObjectFieldTemplateProps) => (
        <Grouped.ObjectFieldTemplate {...props} />
      ),
    };
  }, [sections]);

  const handleSelectProfiles = (profiles: UpdatedProfile[]) => {
    setSelectedProfiles(profiles);
    onProfilesUpdate(profiles);
  };

  useEffect(
    () =>
      setSelectedProfiles(
        updatedProfiles.filter(profile => profile.required === true),
      ),
    [updatedProfiles],
  );

  return (
    <Form
      className={classes.form}
      schema={schema as JSONSchema7}
      onChange={({ formData }) => onFormDataUpdate(formData)}
      formData={formData}
      onSubmit={onPRPreview}
      onError={() => console.log('errors')}
      uiSchema={uiSchema}
      formContext={{
        templates: FormSteps,
        clickedStep,
        setActiveStep,
      }}
      {...UiTemplate}
    >
      <FormStep
        title="Profiles"
        active={activeStep === 'Profiles'}
        clicked={clickedStep === 'Profiles'}
        setActiveStep={setActiveStep}
      >
        <div style={{ display: 'flex', alignItems: 'center' }}>
          <span>Select profiles:&nbsp;</span>
          <MultiSelectDropdown
            items={updatedProfiles}
            onSelectItems={handleSelectProfiles}
          />
        </div>
        <ProfilesList
          selectedProfiles={selectedProfiles}
          onProfilesUpdate={handleSelectProfiles}
        />
        <div className={classes.previewCTA}>
          <Button>Preview PR</Button>
        </div>
      </FormStep>
    </Form>
  );
};

export default TemplateFields;
