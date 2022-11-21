
import { VerifiedUser, Policy } from '@material-ui/icons';
import { ModeWrapper } from './PolicyStyles';

interface IModeProps {
  modeName: string,
  showName?: boolean 
}

function Mode({ modeName, showName }: IModeProps) {
  switch (modeName.toLocaleLowerCase()) {
    case 'audit':
      return ModeTooltip('audit', showName? showName : false, <Policy />);
    case 'admission':
      return ModeTooltip('enforce', showName? showName : false, <VerifiedUser />);
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