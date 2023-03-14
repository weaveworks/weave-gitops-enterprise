import styled from 'styled-components';
import SVGIcon, { IconType } from './WeGOSVGIcon';

type Props = {
  className?: string;
  icon?: StatusIcon;
};

export type StatusIcon = 'error' | 'success' | 'warning' | 'info';

function getType(props: Props, t?: StatusIcon): React.ReactElement | null {
  switch (t) {
    case 'error':
      return <SVGIcon {...props} icon={IconType.AB} />;

    case 'success':
      return <SVGIcon {...props} icon={IconType.Success} />;

    case 'warning':
      return <SVGIcon {...props} icon={IconType.Warning} />;

    default:
      return null;
  }
}

function StatusIcon({ icon, ...rest }: Props) {
  return getType(rest, icon);
}

export default styled(StatusIcon).attrs({ className: StatusIcon.name })``;
