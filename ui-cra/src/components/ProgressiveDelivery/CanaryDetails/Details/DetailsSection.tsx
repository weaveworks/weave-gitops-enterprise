import { ExpandLess, ChevronRight } from '@material-ui/icons';
import {
  Automation,
  Canary,
} from '@weaveworks/progressive-delivery/api/prog/types.pb';
import { formatURL } from '@weaveworks/weave-gitops';
import { useState } from 'react';
import { Link } from 'react-router-dom';
import { getKindRoute } from '../../../../utils/nav';
import { useCanaryStyle } from '../../CanaryStyles';
import { getDeploymentStrategyIcon } from '../../ListCanaries/Table';
import CanaryRowHeader from '../../SharedComponent/CanaryRowHeader';
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
      <CanaryRowHeader rowkey="Cluster" value={canary.clusterName} />
      <CanaryRowHeader rowkey="Namespace" value={canary.namespace} />
      <CanaryRowHeader
        rowkey="Target"
        value={`${canary.targetReference?.kind}/${canary.targetReference?.name}`}
      />
      <CanaryRowHeader
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
      <CanaryRowHeader
        rowkey="Deployment Strategy"
        value={canary.deploymentStrategy}
      >
        <span className={classes.straegyIcon}>
          {getDeploymentStrategyIcon(canary.deploymentStrategy || '')}
        </span>
      </CanaryRowHeader>
      <CanaryRowHeader rowkey="Provider" value={canary.provider} />

      <div className={`${classes.sectionHeaderWrapper} ${classes.cardTitle}`}>
        Status
      </div>
      <DynamicTable obj={restStatus || {}} />
      <div
        className={` ${classes.cardTitle} ${classes.expandableCondition}`}
        onClick={toggleCollapse}
      >
        {!open ? <ExpandLess /> : <ChevronRight />}
        <span className={classes.expandableSpacing}> Conditions</span>
      </div>
      <DynamicTable
        obj={restConditionObj || {}}
        classes={open ? classes.fadeIn : classes.fadeOut}
      />
    </>
  );
};

export default DetailsSection;
