
import { VerifiedUser, Policy } from '@material-ui/icons';
import { ModeWrapper } from './PolicyStyles';

interface IModeProps {
  modeName: string,
  showName?: boolean 
}

function Mode({ modeName, showName = false }: IModeProps) {
  switch (modeName.toLocaleLowerCase()) {
    case 'audit':
      return ModeTooltip('audit', showName, <Policy />);
    case 'admission':
      return ModeTooltip('enforce', showName, <VerifiedUser />);
    default:
      return (
        <ModeWrapper>
          <span>-</span>
        </ModeWrapper>
      );
  }
}

const ModeTooltip = (mode: string, showName: boolean, icon: any) => {
  return (
    <>
      {!showName ? (
        <ModeWrapper>
          <span title={mode}>{icon}</span>
        </ModeWrapper>
      ) : (
        <ModeWrapper>
          {icon}
          <span>{mode}</span>
        </ModeWrapper>
      )}
    </>
  );
};

export default Mode;