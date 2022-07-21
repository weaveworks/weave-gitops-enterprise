import { Alert } from '@material-ui/lab';
import { LoadingPage } from '@weaveworks/weave-gitops';
import { useListFlaggerObjects, CanaryParams } from '../../../contexts/ProgressiveDelivery';
import { ManagedObjectsTable } from './ManagedObjectsTable';

type Props = CanaryParams;

const ListManagedObjects = (props: Props) => {
    const { error, data, isLoading } = useListFlaggerObjects(props);
    return (
        <>
            {isLoading && <LoadingPage />}
            {error && <Alert severity="error">{error.message}</Alert>}
            {data?.objects && <ManagedObjectsTable objects={data.objects} />}
        </>
    );
};

export default ListManagedObjects;
