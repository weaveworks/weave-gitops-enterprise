import * as React from 'react';
import ABIconURL from 'url:../assets/img/ab.svg';
import BlueGreenIconURL from 'url:../assets/img/blue-green.svg';
import CanaryIconURL from 'url:../assets/img/canary.svg';
import ErrorIconURL from 'url:../assets/img/error.svg';
import MirroringIconURL from 'url:../assets/img/mirroring.svg';
import SuccessIconURL from 'url:../assets/img/success.svg';
import WarningIconURL from 'url:../assets/img/warning.svg';

type Props = {
  className?: string;
  icon: SVGIconType;
};

type IconProps = React.HTMLProps<HTMLImageElement>;

export enum SVGIconType {
  Error = 'ErrorIcon',
  Success = 'SuccessIcon',
  Warning = 'WarningIcon',
  AB = 'ABIcon',
  BlueGreen = 'BlueGreenIcon',
  Canary = 'CanaryIcon',
  Mirroring = 'MirroringIcon',
}

function RemoteSVGIcon({ className, icon, ...rest }: Props) {
  let url;

  switch (icon) {
    case SVGIconType.Error:
      url = ErrorIconURL;
      break;

    case SVGIconType.Success:
      url = SuccessIconURL;
      break;

    case SVGIconType.Warning:
      url = WarningIconURL;
      break;

    case SVGIconType.AB:
      url = ABIconURL;
      break;

    case SVGIconType.BlueGreen:
      url = BlueGreenIconURL;
      break;

    case SVGIconType.Canary:
      url = CanaryIconURL;
      break;

    case SVGIconType.Mirroring:
      url = MirroringIconURL;
      break;

    default:
      break;
  }

  return <img className={className} {...rest} src={url} alt={icon} />;
}

export function ErrorIcon(props: IconProps) {
  return <RemoteSVGIcon {...props} icon={SVGIconType.Error} />;
}

export function SuccessIcon(props: IconProps) {
  return <RemoteSVGIcon {...props} icon={SVGIconType.Success} />;
}

export function WarningIcon(props: IconProps) {
  return <RemoteSVGIcon {...props} icon={SVGIconType.Warning} />;
}

export function ABIcon(props: IconProps) {
  return <RemoteSVGIcon {...props} icon={SVGIconType.AB} />;
}

export function BlueGreenIcon(props: IconProps) {
  return <RemoteSVGIcon {...props} icon={SVGIconType.BlueGreen} />;
}

export function CanaryIcon(props: IconProps) {
  return <RemoteSVGIcon {...props} icon={SVGIconType.Canary} />;
}

export function MirroringIcon(props: IconProps) {
  return <RemoteSVGIcon {...props} icon={SVGIconType.Mirroring} />;
}
