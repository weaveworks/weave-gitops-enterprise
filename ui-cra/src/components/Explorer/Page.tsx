import { RouterTab, SubRouterTabs } from '@weaveworks/weave-gitops';
// @ts-ignore
import styled from 'styled-components';
import { Routes } from '../../utils/nav';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';
import AccessRulesDebugger from './AccessRulesDebugger';
import Explorer from './Explorer';

type Props = {
  className?: string;
};

function ExplorerPage({ className }: Props) {
  return (
    <PageTemplate documentTitle="Explorer" path={[{ label: 'Explorer' }]}>
      <ContentWrapper>
        <div className={className}>
          <SubRouterTabs>
            <RouterTab name="Query" path={`${Routes.Explorer}/query`}>
              <Explorer />
            </RouterTab>
            <RouterTab name="Access Rules" path={`${Routes.Explorer}/access`}>
              <AccessRulesDebugger />
            </RouterTab>
          </SubRouterTabs>
        </div>
      </ContentWrapper>
    </PageTemplate>
  );
}

export default styled(ExplorerPage).attrs({ className: ExplorerPage.name })`
  overflow: auto;

  .ExplorerTable {
    flex: 1;
  }
`;
