import { Button } from '@weaveworks/weave-gitops';
import styled from 'styled-components';
import Form from '../Form';
import FormInput from '../FormInput';

type Props = {
  className?: string;
};

type formData = {
  workloadName: string;
  namespace: string;
  clusterName: string;
  image: string;
  branch: string;
  path: string;
  secretName: string;
  sourceRefName: string;
  sourceRefNamespace: string;
};

function ImageAutomationForm({ className }: Props) {
  const handleSubmit = (data: formData) => {
    console.log(data);
  };

  return (
    <div className={className}>
      <Form
        initialState={{ name: '', namespace: '', clusterName: '' }}
        onSubmit={handleSubmit}
      >
        <div>
          <FormInput variant="outlined" label="Name" name="workloadName" />
        </div>
        <div>
          <FormInput label="Namespace" name="namespace" />
        </div>
        <div>
          <FormInput label="Cluster Name" name="clusterName" />
        </div>

        <Button type="submit">Submit</Button>
      </Form>
    </div>
  );
}

export default styled(ImageAutomationForm).attrs({
  className: 'ImageAutomationForm',
})``;
