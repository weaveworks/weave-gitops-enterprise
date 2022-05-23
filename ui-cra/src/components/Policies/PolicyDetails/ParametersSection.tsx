import { Policy, PolicyParam } from '../../../cluster-services/cluster_services.pb';
import { createStyles, makeStyles } from '@material-ui/styles';
import { theme } from '@weaveworks/weave-gitops';

const useStyles = makeStyles(() =>
  createStyles({
    cardTitle: {
      fontWeight: 700,
      fontSize: theme.fontSizes.small,
      color: theme.colors.neutral30,
    },
    body1: {
      fontWeight: 400,
      fontSize: theme.fontSizes.small,
      color: theme.colors.black,
    },
    labelText: {
      fontWeight: 400,
      fontSize: theme.fontSizes.tiny,
      color: theme.colors.neutral30,
    },
    parameterWrapper: {
      border: `1px solid ${theme.colors.neutral20}`,
      boxSizing: 'border-box',
      borderRadius: theme.spacing.xxs,
      padding: theme.spacing.base,
      display: 'flex',
      marginBottom: theme.spacing.base,
      marginTop: theme.spacing.base,
    },
    parameterInfo: {
      display: 'flex',
      alignItems: 'start',
      flexDirection: 'column',
      width: '100%',
    },
    chip: {
      background: theme.colors.neutral10,
      borderRadius: theme.spacing.xxs,
      padding: `${theme.spacing.xxs} ${theme.spacing.xs}`,
      fontWeight: 400,
      fontSize: theme.fontSizes.tiny,
    },
  }),
);
function ParametersSection({ parameters }: Policy) {
  const classes = useStyles();
  const parseValue = (parameter: PolicyParam) => {
    switch (parameter.type) {
      case 'boolean':
        return parameter.value.value ? 'true' : 'false';
      case 'array':
        return parameter.value.value.join(', ');
      case 'string':
        return parameter.value.value;
      case 'integer':
        return parameter.value.value.toString();
    }
  };
  return (
    <>
      <div>
        <span className={classes.cardTitle}>Parameters Definition</span>
        {parameters?.map((parameter: PolicyParam) => (
          <div key={parameter.name} className={classes.parameterWrapper}>
            <div className={classes.parameterInfo}>
              <span className={classes.labelText}>Parameter Name</span>
              <span className={classes.body1}>{parameter.name}</span>
            </div>
            <div className={classes.parameterInfo}>
              <span className={classes.labelText}>Parameter Type</span>
              <span className={classes.body1}>{parameter.type}</span>
            </div>
            <div className={classes.parameterInfo}>
              <span className={classes.labelText}>Value</span>
              <span className={classes.body1}>
                {parameter.value ? (
                  parseValue(parameter)
                ) : (
                  <div className={classes.chip}>undefined</div>
                )}
              </span>
            </div>
            <div className={classes.parameterInfo}>
              <span className={classes.labelText}>Required</span>
              <span className={classes.body1}>
                {parameter.required ? 'True' : 'False'}
              </span>
            </div>
          </div>
        ))}
      </div>
    </>
  );
}

export default ParametersSection;
