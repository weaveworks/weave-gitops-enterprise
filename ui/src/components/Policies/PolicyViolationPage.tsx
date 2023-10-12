import {
  FluxObject,
  Kind,
  V2Routes,
  ViolationDetails,
  formatURL,
} from '@weaveworks/weave-gitops';
import { Breadcrumb } from '@weaveworks/weave-gitops/ui/components/Breadcrumbs';
import { PolicyValidation } from '@weaveworks/weave-gitops/ui/lib/api/core/core.pb';
import styled from 'styled-components';
import { useGetPolicyValidationDetails } from '../../contexts/PolicyViolations';
import { Routes } from '../../utils/nav';
import { Page } from '../Layout/App';
import { NotificationsWrapper } from '../Layout/NotificationsWrapper';

const getPath = (kind?: string, violation?: PolicyValidation): Breadcrumb[] => {
  if (!violation) return [{ label: '' }];
  const { name, entity, namespace, clusterName, policyId } = violation;
  if (!kind) {
    return [{ label: 'Policies', url: `${Routes.Policies}/enforcement` }];
  }

  if (kind === Kind.Policy) {
    const policyUrl = formatURL(`${V2Routes.PolicyDetailsPage}/violations`, {
      id: policyId,
      clusterName,
      name,
    });
    return [
      { label: 'Policies', url: V2Routes.Policies },
      { label: name || '', url: policyUrl },
    ];
  }
  const entityUrl = formatURL(
    kind === Kind.Kustomization
      ? `${V2Routes.Kustomization}/violations`
      : `${V2Routes.HelmRelease}/violations`,
    {
      name: entity,
      namespace: namespace,
      clusterName: clusterName,
    },
  );
  return [
    { label: 'Applications', url: V2Routes.Automations },
    { label: entity || '', url: entityUrl },
  ];
};

interface Props {
  id: string;
  name: string;
  clusterName?: string;
  className?: string;
  kind?: string;
}

const PolicyViolationPage = ({ id, name, clusterName, kind }: Props) => {
  const { data, isLoading } = useGetPolicyValidationDetails({
    validationId: id,
    clusterName,
  });

  const violation = data?.validation;
  const entityObject = new FluxObject({
    payload: violation?.violatingEntity,
  });
  return (
    <Page
      loading={isLoading}
      path={[...getPath(kind, violation), { label: name || '' }]}
    >
      <NotificationsWrapper>
        {violation && (
          <ViolationDetails
            violation={violation}
            entityObject={entityObject}
            kind={kind || ''}
          />
        )}
      </NotificationsWrapper>
    </Page>
  );
};

export default styled(PolicyViolationPage)``;
