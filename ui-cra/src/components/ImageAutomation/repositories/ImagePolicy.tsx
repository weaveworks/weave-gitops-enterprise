import { Box } from '@material-ui/core';
import {
  Flex,
  InfoList,
  Kind,
  KubeStatusIndicator,
  useGetObject,
} from '@weaveworks/weave-gitops';
import { FluxObject } from '@weaveworks/weave-gitops/ui/lib/objects';
import { Header4 } from '../../ProgressiveDelivery/CanaryStyles';

import LoadingWrapper from '../../Workspaces/WorkspaceDetails/Tabs/WorkspaceTabsWrapper';

type Props = {
  name: string;
  namespace: string;
  clusterName: string;
};
const kind = 'ImagePolicy' as Kind; //Kind.ImagePolicy
const ImagePolicy = ({ name, namespace, clusterName }: Props) => {
  const { data, isLoading, error } = useGetObject<FluxObject>(
    name,
    namespace,
    kind,
    clusterName,
    {
      refetchInterval: 5000,
    },
  );
  return (
    <LoadingWrapper loading={isLoading} errorMessage={error?.message}>
      {!!data && (
        <Flex wide tall column>
          <Header4>Policy</Header4>
          <Box margin={2}>
            <KubeStatusIndicator
              short
              conditions={data.conditions || []}
              suspended={data.suspended}
            />
          </Box>
          <InfoList
            items={[
              ['Image Policy', Object.keys(data.obj.spec.policy)[0]],
              ['Order/Range', getValueByKey(data.obj.spec.policy, 'range')],
              ['Kind', kind],
              ['Name', data.name],
              ['Namespace', data.namespace],
            ]}
          />
        </Flex>
      )}
    </LoadingWrapper>
  );
};

export default ImagePolicy;
function getValueByKey(obj: any, key: string): any {
  const policyKey = Object.keys(obj)[0];
  return obj[policyKey][key];
}
