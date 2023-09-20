import { routeTab } from '@weaveworks/weave-gitops/ui/components/KustomizationDetail';
import ImageAutomationUpdatesTable from '../../components/ImageAutomation/updates/ImageAutomationUpdatesTable';
import ImagePoliciesTable from './policies/ImagePoliciesTable';
import ImageRepositoriesTable from './repositories/ImageRepositoriesTable';
import { Flex, RouterTab, SubRouterTabs } from '@weaveworks/weave-gitops';
import { useRouteMatch } from 'react-router-dom';

const ImageAutomation = () => {
  const { path } = useRouteMatch();

  const tabs: Array<routeTab> = [
    {
      name: 'Image Repositories',
      path: `${path}/repositories`,
      component: () => {
        return <ImageRepositoriesTable />;
      },
      visible: true,
    },
    {
      name: 'Image Policies',
      path: `${path}/policies`,
      component: () => {
        return <ImagePoliciesTable />;
      },
      visible: true,
    },
    {
      name: 'Image Update Automations',
      path: `${path}/updates`,
      component: () => {
        return <ImageAutomationUpdatesTable />;
      },
      visible: true,
    },
  ];
  return (
    <Flex wide tall column>
      <SubRouterTabs rootPath={tabs[0].path} clearQuery>
        {tabs.map(
          (subRoute, index) =>
            subRoute.visible && (
              <RouterTab name={subRoute.name} path={subRoute.path} key={index}>
                {subRoute.component()}
              </RouterTab>
            ),
        )}
      </SubRouterTabs>
    </Flex>
  );
};

export default ImageAutomation;
