import { VerifiedUser, Policy } from '@material-ui/icons';
import { ModeWrapper } from './PolicyStyles';

function Mode({ modeName }: { modeName: string }) {
  switch (modeName.toLocaleLowerCase()) {
    case 'audit':
      return (
        <ModeWrapper>
          <Policy />
          <span>{modeName}</span>
        </ModeWrapper>
      );
    case 'admission':
      return (
        <ModeWrapper>
          <VerifiedUser />
          <span>Enforce</span>
        </ModeWrapper>
      );
    default:
      return (
        <ModeWrapper>
          <span>{modeName}</span>
        </ModeWrapper>
      );
  }
}

export default Mode;
