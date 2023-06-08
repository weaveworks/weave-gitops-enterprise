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
import { toFilterQueryString } from '../../../utils/FilterQueryString';
import { NotificationsWrapper } from '../../Layout/NotificationsWrapper';
import ImageAutomationDetails from '../ImageAutomationDetails';
import { Page } from '../../Layout/App';

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

  const filtersValues = toFilterQueryString([
    { key: 'imageRepositoryRef', value: name },
  ]);
  const rootPath = V2Routes.ImageAutomationRepositoryDetails;

  return (
    <Page
      loading={isLoading}
      path={[
        { label: 'Image Repositories', url: V2Routes.ImageRepositories },
        { label: name },
      ]}
    >
      <NotificationsWrapper>
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
      </NotificationsWrapper>
    </Page>
  );
}

export default styled(ImageAutomationRepoDetails)``;
