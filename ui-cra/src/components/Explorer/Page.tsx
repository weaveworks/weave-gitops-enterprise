import { Page, RouterTab, SubRouterTabs } from '@weaveworks/weave-gitops';
// @ts-ignore
import styled from 'styled-components';
import { Routes } from '../../utils/nav';
import AccessRulesDebugger from './AccessRulesDebugger';
import Explorer from './Explorer';

type Props = {
  className?: string;
};

function ExplorerPage({ className }: Props) {
  return (
    <Page path={[{ label: 'Explorer' }]}>
      <div className={className}>
        <SubRouterTabs rootPath={`${Routes.Explorer}/query`}>
          <RouterTab name="Query" path={`${Routes.Explorer}/query`}>
            <Explorer />
          </RouterTab>
          <RouterTab name="Access Rules" path={`${Routes.Explorer}/access`}>
            <AccessRulesDebugger />
          </RouterTab>
        </SubRouterTabs>
      </div>
    </Page>
  );
}

export default styled(ExplorerPage).attrs({ className: ExplorerPage.name })`
  overflow: auto;

  .ExplorerTable {
    flex: 1;
  }
`;
