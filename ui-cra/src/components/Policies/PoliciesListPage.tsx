import { NotificationsWrapper } from '../Layout/NotificationsWrapper';
import { Page } from '../Layout/App';
import { useListPolicies } from '../../contexts/PolicyViolations';
import {
  Flex,
  PolicyTable,
  PolicyViolationsList,
  RouterTab,
  SubRouterTabs,
} from '@weaveworks/weave-gitops';
import { useRouteMatch } from 'react-router-dom';

const Policies = () => {
  const { data, isLoading } = useListPolicies({});
  const { path } = useRouteMatch();

  return (
    <Page loading={isLoading} path={[{ label: 'Policies' }]}>
      <NotificationsWrapper errors={data?.errors}>
        <SubRouterTabs rootPath={`${path}/`}>
          <RouterTab name="Policies" path={`${path}/list/`}>
            <>
              {data?.policies && (
                <Flex wide id="policy-list">
                  <PolicyTable policies={data.policies} />
                </Flex>
              )}
            </>
          </RouterTab>
          <RouterTab name="Audit Violations" path={`${path}/auditViolations/`}>
            <div>violations</div>
          </RouterTab>
          <RouterTab
            name="Enforcement Events"
            path={`${path}/enforcementEvent/`}
          >
            <PolicyViolationsList req={{}} />
          </RouterTab>
        </SubRouterTabs>
      </NotificationsWrapper>
    </Page>
  );
};

export default Policies;
