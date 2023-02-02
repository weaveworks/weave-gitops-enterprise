import { Flex, RouterTab, SubRouterTabs } from '@weaveworks/weave-gitops';
import { routeTab } from '@weaveworks/weave-gitops/ui/components/KustomizationDetail';
import { useRouteMatch } from 'react-router-dom';
import ImageRepositoriesTable from '../../components/ImageAutomation/repositories/ImageRepositoriesTable';
import ImageAutomationUpdatesTable from '../../components/ImageAutomation/updates/ImageAutomationUpdatesTable';

const ImageAutomation = () => {
  const { path } = useRouteMatch();

  const tabs: Array<routeTab> = [
    {
      name: 'Image Update Automations',
      path: `${path}/updates`,
      component: () => {
        return <ImageAutomationUpdatesTable />;
      },
      visible: true,
    },
    {
      name: 'Image Repositories',
      path: `${path}/repositories`,
      component: () => {
        return <ImageRepositoriesTable />;
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
