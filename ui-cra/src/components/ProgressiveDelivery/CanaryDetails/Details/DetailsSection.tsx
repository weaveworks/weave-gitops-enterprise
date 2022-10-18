import { ExpandLess, ChevronRight } from '@material-ui/icons';
import {
  Automation,
  Canary,
} from '@weaveworks/progressive-delivery/api/prog/types.pb';
import { formatURL, Link } from '@weaveworks/weave-gitops';
import { useState } from 'react';
import { getKindRoute } from '../../../../utils/nav';
import { useCanaryStyle } from '../../CanaryStyles';
import { getDeploymentStrategyIcon } from '../../ListCanaries/Table';
import RowHeader from '../../../RowHeader';
import DynamicTable from '../../SharedComponent/DynamicTable';

const DetailsSection = ({
  canary,
  automation,
}: {
  canary: Canary;
  automation?: Automation;
}) => {
  const classes = useCanaryStyle();
  const [open, setOpen] = useState(true);
  const { conditions, ...restStatus } = canary?.status || { conditions: [] };
  const { lastTransitionTime, ...restConditionObj } = conditions![0] || {
    lastTransitionTime: '',
  };
  const toggleCollapse = () => {
    setOpen(!open);
  };
  return (
    <>
      <RowHeader rowkey="Cluster" value={canary.clusterName} />
      <RowHeader rowkey="Namespace" value={canary.namespace} />
      <RowHeader
        rowkey="Target"
        value={`${canary.targetReference?.kind}/${canary.targetReference?.name}`}
      />
      <RowHeader
        rowkey="Application"
        value={
          automation?.kind && automation?.name ? (
            <Link
              to={formatURL(getKindRoute(automation?.kind), {
                name: automation?.name,
                namespace: automation?.namespace,
                clusterName: canary.clusterName,
              })}
            >
              {automation?.kind}/{automation?.name}
            </Link>
          ) : (
            ''
          )
        }
      />
      <RowHeader rowkey="Deployment Strategy" value={undefined}>
        {!!canary.deploymentStrategy && (
          <span className={classes.straegyIcon}>
            {canary.deploymentStrategy}{' '}
            {getDeploymentStrategyIcon(canary.deploymentStrategy)}
          </span>
        )}
      </RowHeader>
      <RowHeader rowkey="Provider" value={canary.provider} />

      <div className={`${classes.sectionHeaderWrapper} ${classes.cardTitle}`}>
        STATUS
      </div>
      <DynamicTable obj={restStatus || {}} />
      <div
        className={` ${classes.cardTitle} ${classes.expandableCondition}`}
        onClick={toggleCollapse}
      >
        {!open ? <ExpandLess /> : <ChevronRight />}
        <span className={classes.expandableSpacing}> CONDITIONS</span>
      </div>
      <DynamicTable
        obj={restConditionObj || {}}
        classes={open ? classes.fadeIn : classes.fadeOut}
      />
    </>
  );
};

export default DetailsSection;
