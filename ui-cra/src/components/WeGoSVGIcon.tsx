import React from 'react';
import styled from 'styled-components';

import ABIcon from '../assets/img/ab.svg';
import Applications from '../assets/img/applications.svg';
import BlueGreenIcon from '../assets/img/blue-green.svg';
import CanaryIcon from '../assets/img/canary.svg';
import Clusters from '../assets/img/clusters.svg';
import ErrorIcon from '../assets/img/error.svg';
import FluxIcon from '../assets/img/flux-icon.svg';
import GitOpsRun from '../assets/img/gitops-run-icon.svg';
import GitOpsSetsIcon from '../assets/img/gitopssets.svg';
import MenuIcon from '../assets/img/menu-burger.svg';
import MirroringIcon from '../assets/img/mirroring.svg';
import Policies from '../assets/img/policies.svg';
import PolicyConfigs from '../assets/img/policyConfigs.svg';
import SecretsIcon from '../assets/img/secrets-Icon.svg';
import SuccessIcon from '../assets/img/success.svg';
import Templates from '../assets/img/templates.svg';
import TerraformLogo from '../assets/img/terraform-logo.svg';
import WarningIcon from '../assets/img/warning.svg';
import WorkspacesIcon from '../assets/img/Workspace-Icon.svg';

type Props = {
  className?: string;
  icon?: IconType;
} & React.HTMLProps<HTMLImageElement>;

export enum IconType {
  AB = 'ab',
  BlueGreen = 'blue-green',
  Mirroring = 'mirror',
  Canary = 'canary',
  Error = 'error',
  Success = 'success',
  Warning = 'warning',
  Applications = 'applications',
  Clusters = 'clusters',
  Flux = 'flux',
  GitOpsRun = 'gitops-run',
  GitOpsSetsIcon = 'gitopssets',
  Policies = 'policies',
  PolicyConfigs = 'policy-configs',
  SecretsIcon = 'secrets',
  Templates = 'templates',
  TerraformLogo = 'terraform',
  Workspaces = 'workspace',
  MenuIcon = 'menu-icon',
}

function getType(t?: IconType) {
  switch (t) {
    case IconType.AB:
      return ABIcon;

    case IconType.BlueGreen:
      return BlueGreenIcon;

    case IconType.Mirroring:
      return MirroringIcon;

    case IconType.Canary:
      return CanaryIcon;

    case IconType.Error:
      return ErrorIcon;

    case IconType.Success:
      return SuccessIcon;

    case IconType.Warning:
      return WarningIcon;

    case IconType.Applications:
      return Applications;

    case IconType.Clusters:
      return Clusters;

    case IconType.Flux:
      return FluxIcon;

    case IconType.GitOpsRun:
      return GitOpsRun;

    case IconType.GitOpsSetsIcon:
      return GitOpsSetsIcon;

    case IconType.Policies:
      return Policies;

    case IconType.PolicyConfigs:
      return PolicyConfigs;

    case IconType.SecretsIcon:
      return SecretsIcon;

    case IconType.Templates:
      return Templates;

    case IconType.TerraformLogo:
      return TerraformLogo;

    case IconType.Workspaces:
      return WorkspacesIcon;

    case IconType.MenuIcon:
      return MenuIcon;

    default:
      return '';
  }
}

function WeGoSVGIcon({ icon, className }: Props) {
  return <img src={getType(icon)} className={className} />;
}

export default styled(WeGoSVGIcon).attrs({ className: WeGoSVGIcon.name })``;
