import { Button } from '@weaveworks/weave-gitops';
import { useState } from 'react';
import { InputDebounced } from '../../../utils/form';
import ListClusters from '../../Secrets/Shared/ListClusters';

interface TargetProps {
  clusterName?: string;
  namespace?: string;
  clusterNamespace?: string;
  environment?: string;
}

const AddTarget = ({
  addTarget,
}: {
  addTarget: (newTarget: TargetProps) => void;
}) => {
  const [targetForm, setTarget] = useState<TargetProps>({});

  const handleAddTarget = () => {
    addTarget(targetForm);
    setTarget({});
  };

  return (
    <>
      <ListClusters
        value={targetForm?.clusterName || ''}
        validateForm={false}
        handleFormData={(val: any) => {
          const [ns, cluster] = val.split('/');

          setTarget(f => ({
            ...f,
            clusterName: !cluster ? ns : cluster,
            clusterNamespace: cluster ? ns : '',
            namespace: '',
          }));
        }}
      />
      <InputDebounced
        required
        name="namespace"
        label="Namespace"
        value={targetForm.namespace || ''}
        handleFormData={val => {
          setTarget(f => ({
            ...f,
            namespace: val,
          }));
        }}
      />
      <Button onClick={handleAddTarget}>Add Target</Button>
    </>
  );
};

export default AddTarget;
