
import { VerifiedUser, Policy } from '@material-ui/icons';
import { ModeWrapper } from './PolicyStyles';

interface IModeProps {
  modeName: string,
  showName?: boolean 
}
const capitalizeFirstLetter =( strToCapitalize:string) =>  strToCapitalize.charAt(0).toUpperCase() + strToCapitalize.slice(1);

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
          <span title={capitalizeFirstLetter(mode)}>{icon}</span>
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