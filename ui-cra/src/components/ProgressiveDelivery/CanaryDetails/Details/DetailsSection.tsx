import { Table, TableBody } from '@material-ui/core';
import { ExpandLess, ExpandMore } from '@material-ui/icons';
import {
  Automation,
  Canary,
} from '@weaveworks/progressive-delivery/api/prog/types.pb';
import { formatURL } from '@weaveworks/weave-gitops';
import  { useState } from 'react';
import { Link } from 'react-router-dom';
import { getKindRoute } from '../../../../utils/nav';
import { useCanaryStyle } from '../../CanaryStyles';
import { getDeploymentStrategyIcon } from '../../ListCanaries/Table';
import CanaryRowHeader, {
  KeyValueRow,
} from '../../SharedComponent/CanaryRowHeader';

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
      <Table size="small">
        <TableBody>
          {Object.entries(restStatus || {}).map((entry, index) => (
            <KeyValueRow entryObj={entry} key={index} />
          ))}
        </TableBody>
      </Table>

      <div
        className={` ${classes.cardTitle} ${classes.expandableCondition}`}
        onClick={toggleCollapse}
      >
        {!open ? <ExpandLess /> : <ExpandMore />}
        <span className={classes.expandableSpacing}> Conditions</span>
      </div>
      <Table size="small" className={open ? classes.fadeIn : classes.fadeOut}>
        <TableBody>
          {Object.entries(restConditionObj).map((entry, index) => (
            <KeyValueRow entryObj={entry} key={index}></KeyValueRow>
          ))}
        </TableBody>
      </Table>
    </>
  );
};

export default DetailsSection;
