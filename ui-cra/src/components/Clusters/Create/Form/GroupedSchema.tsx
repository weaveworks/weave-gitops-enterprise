// Adapted from : https://codesandbox.io/s/0y7787xp0l?file=/src/index.js:1507-1521

import { ObjectFieldTemplateProps } from '@rjsf/core';
import React, { Dispatch, ReactNode } from 'react';

export function ObjectFieldTemplate(props: ObjectFieldTemplateProps) {
  return (
    <>
      {doGrouping({
        formContext: props.formContext,
        properties: props.properties,
        groups: props.uiSchema['ui:groups'],
      })}
    </>
  );
}

const EXTRANEOUS = Symbol('EXTRANEOUS');

// children: doGrouping({
//   properties,
//   groups: field,
//   formContext,
// }),

function doGrouping({
  properties,
  groups,
  formContext,
}: {
  properties: ObjectFieldTemplateProps['properties'];
  formContext: ObjectFieldTemplateProps['formContext'];
  groups: string | object;
}) {
  if (!Array.isArray(groups)) {
    return properties?.map((property, index) => {
      return <div key={index}>{property.content}</div>;
    });
  }
  const mapped = groups.map((g, index) => {
    if (typeof g === 'string') {
      const found = properties?.filter(property => property.name === g);
      if (found?.length === 1) {
        const el = found[0];
        return el.content;
      }
      return EXTRANEOUS;
    } else if (typeof g === 'object') {
      const { templates, activeStep, setActiveStep, clickedStep } = formContext;
      const GroupComponent = templates
        ? templates[g['ui:template']]
        : DefaultTemplate;
      console.log(g);

      // keys(g)
      // 1.Cluster
      // 2.AWSCluster
      // 3.KubeadmControlPlane
      // 4.AWSMachineTemplate
      // 5.MachineDeployment
      // 6.AWSMachineTemplate
      // 7.KubeadmConfigTemplate

      const _properties = Object.keys(g).reduce(
        (
          acc: {
            name: string;
            active: boolean;
            clicked: boolean;
            setActiveStep: Dispatch<React.SetStateAction<string | undefined>>;
            children: ReactNode;
          }[],
          key: string,
        ) => {
          const field = g[key];
          // the key is the group title

          // the field is what's under the group
          // ['CLUSTER_NAME'] OR ['AWS_REGION', 'AWS_SSH_KEY_NAME', 'CLUSTER_NAME']

          if (key.startsWith('ui:')) return acc;
          if (!Array.isArray(field)) return acc;

          // the key is the next one so wont be in the accumulator.
          // go through the accumulator's keys and children  key.children (array).key
          // if you find any of the parameters that field has then move on. if not, add a key.props.visible => true

          // example acc
          // [{…}]
          // 0:
          // active: false
          // children: (2) [{…}, {…}]
          // clicked: false
          // name: "1.Cluster"
          // setActiveStep: ƒ ()
          // [[Prototype]]: Object

          // children: Array(2)
          // 0:
          // $$typeof: Symbol(react.element)
          // key: "CLUSTER_NAME"
          // props: {name: 'CLUSTER_NAME', required: true, schema: {…}, uiSchema: {…}, errorSchema: {…}, …}
          // ref: null
          // type: ƒ SchemaField()
          // _owner: FiberNode {tag: 1, key: null, stateNode: ObjectField, elementType: ƒ, type: ƒ, …}
          // _store: {validated: false}
          // _self: null
          // _source: null
          // [[Prototype]]: Object

          const elemInAcc = acc.find(element => element.name === key);

          console.log('acc', acc);
          console.log('key', key);
          console.log('field or the key s fields', field);

          console.log(elemInAcc);

          return [
            ...acc,
            {
              name: key,
              active: key === activeStep,
              clicked: key === clickedStep,
              setActiveStep,
              children: doGrouping({
                formContext,
                properties,
                groups: field,
              }),
            },
          ];
        },
        [],
      );

      return <GroupComponent key={index} properties={_properties} />;
    }
    throw new Error('Invalid object type: ' + typeof g + ' ' + g);
  });
  return mapped;
}

function DefaultTemplate(props: {
  properties: ObjectFieldTemplateProps['properties'];
}) {
  return props?.properties?.map((p, index) => (
    <div key={index}>{p.content}</div>
  ));
}
