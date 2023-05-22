import { Flex, RouterTab, SubRouterTabs } from '@weaveworks/weave-gitops';
import { routeTab } from '@weaveworks/weave-gitops/ui/components/KustomizationDetail';
import ImageAutomationUpdatesTable from '../../components/ImageAutomation/updates/ImageAutomationUpdatesTable';
import ImagePoliciesTable from './policies/ImagePoliciesTable';
import ImageRepositoriesTable from './repositories/ImageRepositoriesTable';

const ImageAutomation = () => {
  const tabs: Array<routeTab> = [
    {
      name: 'Image Repositories',
      path: `/repositories`,
      component: () => {
        return <ImageRepositoriesTable />;
      },
      visible: true,
    },
    {
      name: 'Image Policies',
      path: `/policies`,
      component: () => {
        return <ImagePoliciesTable />;
      },
      visible: true,
    },
    {
      name: 'Image Update Automations',
      path: `/updates`,
      component: () => {
        return <ImageAutomationUpdatesTable />;
      },
      visible: true,
    },
  ];
  return (
    <Flex wide tall column>
      <SubRouterTabs clearQuery>
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
