import { createStyles, makeStyles } from '@material-ui/styles';
import styled from 'styled-components';
import { CallReceived, CallMade, Remove } from '@material-ui/icons';
import { theme } from '@weaveworks/weave-gitops';

const useStyles = makeStyles(() =>
  createStyles({
    severityIcon: {
      fontSize: theme.fontSizes.small,
      marginRight: theme.spacing.xxs,
    },
    severityLow: {
      color: '#006B8E',
    },
    severityMedium: {
      color: '#8A460A',
    },
    severityHigh: {
      color: '#9F3119',
    },
  }),
);

const FlexStart = styled.div`
  display: flex;
  align-items: center;
  justify-content: start;
`;

function Severity({ severity }: { severity: string }) {
  const classes = useStyles();
  switch (severity.toLocaleLowerCase()) {
    case 'low':
      return (
        <FlexStart>
          <CallReceived
            className={`${classes.severityIcon} ${classes.severityLow}`}
          />
          <span>{severity}</span>
        </FlexStart>
      );
    case 'high':
      return (
        <FlexStart>
          <CallMade
            className={`${classes.severityIcon} ${classes.severityHigh}`}
          />
          <span>{severity}</span>
        </FlexStart>
      );
    case 'medium':
      return (
        <FlexStart>
          <Remove
            className={`${classes.severityIcon} ${classes.severityMedium}`}
          />
          <span>{severity}</span>
        </FlexStart>
      );

    default:
      return <span>{severity}</span>;
  }
}

export default Severity;
