import { Alert } from '@material-ui/lab';
import { LoadingPage } from '@weaveworks/weave-gitops';
import { useListFlaggerObjects, CanaryParams } from '../../../../contexts/ProgressiveDelivery';
import { ManagedObjectsTable } from './ManagedObjectsTable';
import { theme } from '@weaveworks/weave-gitops';
import AlertTitle from '@material-ui/lab/AlertTitle';
import { createStyles, makeStyles } from '@material-ui/styles';
import { ListItem } from '@material-ui/core';
import { useState } from 'react';
import { ListError } from '../../../../cluster-services/cluster_services.pb';

type Props = CanaryParams;
const useStyles = makeStyles(() =>
  createStyles({
    alertWrapper: {
      marginTop: theme.spacing.medium,
      marginRight: theme.spacing.small,
      marginBottom: 0,
      marginLeft: theme.spacing.small,
      paddingRight: theme.spacing.medium,
      paddingLeft: theme.spacing.medium,
      borderRadius: theme.spacing.xs,
    },
    warning: {
      backgroundColor: theme.colors.feedbackLight,
    },
  }),
);

const ListManagedObjects = (props: Props) => {
    const classes = useStyles();
    // const [errors, setErrors] = useState<ListError[] | undefined>();

    const { error, data, isLoading } = useListFlaggerObjects(props);


    let errors:ListError[] = [{
        clusterName: "management",
        message: "request is forbidden trafficsplits.split.smi-spec.io is forbidden: User \"wego-admin\" cannot list resource \"trafficsplits\" in API group \"split.smi-spec.io\" at the cluster scope",
        namespace: "default"
    }]
    // setErrors(data?.errors);

    return (
        <>
            {!!(errors && errors.length) && (
                <Alert className={classes.alertWrapper} severity="error">
                <AlertTitle>
                    There was a problem retrieving results from some clusters:
                </AlertTitle>
                {errors?.map((item: ListError) => (
                    <ListItem key={item.clusterName}>
                    - Cluster {item.clusterName} {item.message}
                    </ListItem>
                ))}
                </Alert>
            )}
            {isLoading && <LoadingPage />}
            {error && <Alert severity="error">{error.message}</Alert>}
            {data?.objects &&
                    <ManagedObjectsTable objects={data.objects} />
            }
        </>
    );
};

export default ListManagedObjects;
