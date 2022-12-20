import { RouterTab, SubRouterTabs } from '@weaveworks/weave-gitops';
import { Pipeline } from '../../../api/pipelines/types.pb';
import { useGetPipeline } from '../../../contexts/Pipelines';
import { Routes } from '../../../utils/nav';
import { ContentWrapper } from '../../Layout/ContentWrapper';
import { PageTemplate } from '../../Layout/PageTemplate';
import YamlView from '../../YamlView';

import { EditButton } from './../../../components/Templates/Edit/EditButton';
import { ListError } from '@weaveworks/progressive-delivery/api/prog/types.pb';
import { usePipelineStyles } from './styles';
import Workloads from './Workloads';

const mappedErrors = (
  errors: Array<string>,
  namespace: string,
): Array<ListError> => {
  return errors.map(err => ({
    message: err,
    namespace,
  }));
};
interface Props {
  name: string;
  namespace: string;
}

const PipelineDetails = ({ name, namespace }: Props) => {
  const { isLoading, data } = useGetPipeline({
    name,
    namespace,
  });

  const classes = usePipelineStyles();
  const path = `/applications/pipelines/details`;

  return (
    <PageTemplate
      documentTitle="Pipeline Details"
      path={[
        {
          label: 'Applications',
          url: Routes.Applications,
        },
        {
          label: 'Pipelines',
          url: Routes.Pipelines,
        },
        {
          label: name,
        },
      ]}
    >
      <ContentWrapper
        loading={isLoading}
        errors={mappedErrors(data?.errors || [], namespace)}
      >
        <EditButton
          className={classes.editButton}
          resource={data?.pipeline || ({} as Pipeline)}
        />
        <SubRouterTabs rootPath={`${path}/status`}>
          <RouterTab name="Status" path={`${path}/status`}>
            <Workloads pipeline={data?.pipeline || ({} as Pipeline)} />
          </RouterTab>
          <RouterTab name="Yaml" path={`${path}/yaml`}>
            <YamlView
              kind="Pipeline"
              yaml={data?.pipeline?.yaml || ''}
              object={data?.pipeline || {}}
            />
          </RouterTab>
        </SubRouterTabs>
      </ContentWrapper>
    </PageTemplate>
  );
};

export default PipelineDetails;
