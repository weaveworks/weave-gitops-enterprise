import React, { FC, useCallback, useMemo, useState } from 'react';
import useTemplates from '../../../contexts/Templates';
import useClusters from '../../../contexts/Clusters';
import { PageTemplate } from '../../Layout/PageTemplate';
import { SectionHeader } from '../../Layout/SectionHeader';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import Divider from '@material-ui/core/Divider';
import { useHistory } from 'react-router-dom';
import Form from '@rjsf/material-ui';
import { JSONSchema7 } from 'json-schema';
import { makeStyles, createStyles } from '@material-ui/core/styles';
import { Button } from 'weaveworks-ui-components';

const useStyles = makeStyles(() =>
  createStyles({
    form: {},
    divider: {
      marginTop: 10,
      marginBottom: 10,
    },
    buttonBox: {
      display: 'flex',
      justifyContent: 'flex-end',
      paddingTop: '16px',
    },
  }),
);

const AddCluster: FC = () => {
  const classes = useStyles();
  const { activeTemplate, setActiveTemplate } = useTemplates();
  const { addCluster, count } = useClusters();
  const [formData] = useState({});
  const history = useHistory();

  const handleAddCluster = useCallback(
    (event: { formData: any }) => {
      addCluster({ ...formData, ...event.formData });
      setActiveTemplate(null);
      history.push('/clusters');
    },
    [addCluster, formData, history, setActiveTemplate],
  );
  const required = useMemo(() => {
    return activeTemplate?.parameters?.map(param => param.name);
  }, [activeTemplate]);

  const parameters = useMemo(() => {
    return (
      activeTemplate?.parameters?.map(param => {
        const name = param.name;
        return {
          [name]: { type: 'string', title: `${name}` },
        };
      }) || []
    );
  }, [activeTemplate]);

  const properties = useMemo(() => {
    return Object.assign({}, ...parameters);
  }, [parameters]);

  const schema: JSONSchema7 = useMemo(() => {
    return {
      title: 'Cluster',
      type: 'object',
      required,
      properties,
    };
  }, [properties, required]);

  return useMemo(() => {
    return (
      <PageTemplate documentTitle="WeGo · Create new cluster">
        <SectionHeader
          path={[
            { label: 'Clusters', url: '/', count },
            { label: 'Create new cluster' },
          ]}
        />
        <ContentWrapper>
          Template: {activeTemplate?.name}
          <Divider className={classes.divider} />
          <Form
            className={classes.form}
            schema={schema as JSONSchema7}
            onChange={() => console.log('changed')}
            formData={formData}
            onSubmit={handleAddCluster}
            onError={() => console.log('errors')}
          >
            <div className={classes.buttonBox}>
              <Button>Create a Pull Request</Button>
            </div>
          </Form>
        </ContentWrapper>
      </PageTemplate>
    );
  }, [
    count,
    activeTemplate?.name,
    classes.buttonBox,
    classes.divider,
    classes.form,
    formData,
    handleAddCluster,
    schema,
  ]);
};

export default AddCluster;
