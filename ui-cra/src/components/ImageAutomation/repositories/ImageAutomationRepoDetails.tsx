import {
  Button,
  Interval,
  Kind,
  Link,
  useGetObject,
  V2Routes,
} from '@weaveworks/weave-gitops';
import { ImageRepository } from '@weaveworks/weave-gitops/ui/lib/objects';
import styled from 'styled-components';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';

import ImageAutomationDetails from '../ImageAutomationDetails';

type Props = {
  className?: string;
  name: string;
  namespace: string;
  clusterName: string;
};

function ImageAutomationRepoDetails({ name, namespace, clusterName }: Props) {
  const { data, isLoading } = useGetObject<ImageRepository>(
    name,
    namespace,
    Kind.ImageRepository,
    clusterName,
    {
      refetchInterval: 30000,
    },
  );

  const filtersValues = encodeURIComponent(`imageRepositoryRef: ${name}`);
  const rootPath = V2Routes.ImageAutomationRepositoryDetails;
  return (
    <PageTemplate
      documentTitle={name}
      path={[
        { label: 'Image Repositories', url: V2Routes.ImageRepositories },
        { label: name },
      ]}
    >
      <ContentWrapper loading={isLoading}>
        {data && (
          <ImageAutomationDetails
            data={data}
            kind={Kind.ImageRepository}
            infoFields={[
              ['Kind', Kind.ImageRepository],
              ['Namespace', data?.namespace],
              ['Namespace', data?.clusterName],
              [
                'Image',
                <Link newTab={true} to={data.obj?.spec?.image}>
                  {data.obj?.spec?.image}
                </Link>,
              ],
              ['Interval', <Interval interval={data.interval} />],
              ['Tag Count', data.tagCount],
            ]}
            rootPath={rootPath}
          >
            <Button>
              <Link to={`/image_automation/policies?filters=${filtersValues}`}>
                Go To Image Policy
              </Link>
            </Button>
          </ImageAutomationDetails>
        )}
      </ContentWrapper>
    </PageTemplate>
  );
}

export default styled(ImageAutomationRepoDetails)``;
