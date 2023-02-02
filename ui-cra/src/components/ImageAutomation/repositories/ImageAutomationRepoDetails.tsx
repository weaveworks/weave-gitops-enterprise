import { Interval, Kind, useGetObject } from '@weaveworks/weave-gitops';
import { ListError } from '@weaveworks/weave-gitops/ui/lib/api/core/core.pb';
import { FluxObject } from '@weaveworks/weave-gitops/ui/lib/objects';

import styled from 'styled-components';
import { Routes } from '../../../utils/nav';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';
import { LinkTag } from '../../ProgressiveDelivery/CanaryStyles';

import ImageAutomationDetails from '../ImageAutomationDetails';
import ImagePolicy from './ImagePolicy';

type Props = {
  className?: string;
  name: string;
  namespace: string;
  clusterName: string;
};
const kind = 'ImageRepository' as Kind; //Kind.ImageRepository
function ImageAutomationRepoDetails({
  className,
  name,
  namespace,
  clusterName,
}: Props) {
  const { data, isLoading, error } = useGetObject<FluxObject>(
    name,
    namespace,
    kind,
    clusterName,
    {
      refetchInterval: 5000,
    },
  );
  const rootPath = Routes.ImageAutomationRepositoriesDetails;
  return (
    <PageTemplate
      documentTitle="Image Automation Updates"
      path={[
        { label: 'Image Automation', url: Routes.ImageAutomation },
        { label: name },
      ]}
    >
      <ContentWrapper loading={isLoading} errors={[error as ListError]}>
        {!!data && (
          <ImageAutomationDetails
            data={data}
            kind={kind}
            infoFields={[
              ['Kind', kind],
              ['Namespace', data.namespace],
              [
                'Image',
                <LinkTag newTab={true} to={data.obj?.spec?.image}>
                  {data.obj?.spec?.image}
                </LinkTag>,
              ],
              ['Interval', <Interval interval={data.interval} />],
              ['Tag Count', data.obj?.status?.lastScanResult?.tagCount],
            ]}
            rootPath={rootPath}
          >
            <ImagePolicy
              clusterName={clusterName}
              name={name}
              namespace={namespace}
            />
          </ImageAutomationDetails>
        )}
      </ContentWrapper>
    </PageTemplate>
  );
}

export default styled(ImageAutomationRepoDetails)``;
