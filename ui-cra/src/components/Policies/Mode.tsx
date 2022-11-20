import { VerifiedUser, Policy } from '@material-ui/icons';
import { ModeWrapper } from './PolicyStyles';

export function ListMode({ modeName }: { modeName: string }) {
  switch (modeName.toLocaleLowerCase()) {
    case 'audit':
      return (
        <ModeWrapper>
          <Policy />
          <span>{modeName}</span>
        </ModeWrapper>
      );
    case 'enforce':
      return (
        <ModeWrapper>
          <VerifiedUser />
          <span>{modeName}</span>
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

function Mode({ modeName }: { modeName: string }) {
  const modes = modeName.split(' ');
  return (
    <>
      {modes.map((mode, index) => (
        <ListMode modeName={mode} key={index} />
      ))}
    </>
  );
}

export function mapPolicyMode(modes: string[]): string {
 return modes?.sort().reduce((prev, next) => {
    const nextVal = next === 'admission' ? 'enforce' : next;
    console.log('prev',prev,'next', nextVal)
    return prev ? prev + ' ' + nextVal :nextVal;
  }, '')
}

export default Mode;
