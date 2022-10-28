import { CircularProgress } from '@material-ui/core';
import { Alert } from '@material-ui/lab';
import {
  Flex,
  Kind,
  KubeStatusIndicator,
  useGetObject,
} from '@weaveworks/weave-gitops';
import { HelmRelease } from '@weaveworks/weave-gitops/ui/lib/objects';
import styled from 'styled-components';

interface Props {
  className?: string;
  name: string;
  namespace: string;
  clusterName: string;
}

function WorkloadStatus({ className, name, namespace, clusterName }: Props) {
  const { data, isLoading, error } = useGetObject<HelmRelease>(
    name,
    namespace,
    Kind.HelmRelease,
    clusterName,
  );

  return (
    <div className={className}>
      <Flex align>
        <div style={{ marginRight: 4 }}>
          {isLoading ? (
            <CircularProgress size={12} />
          ) : (
            <KubeStatusIndicator
              short
              conditions={data?.obj?.status?.conditions}
            />
          )}
        </div>
        <div>{name}</div>
      </Flex>
      {error && <Alert severity="error">{error.message}</Alert>}
    </div>
  );
}

export default styled(WorkloadStatus).attrs({ className: 'WorkloadStatus' })`
  font-size: ${props => props.theme.fontSizes.large};

  /* Hack to hide the text; we only want the icon here */
  ${KubeStatusIndicator} span {
    display: none;
  }
`;
