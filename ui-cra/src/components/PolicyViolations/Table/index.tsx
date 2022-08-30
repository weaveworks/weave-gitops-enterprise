import { FC } from 'react';
import {
  ListPolicyValidationsRequest,
  PolicyValidation,
} from '../../../cluster-services/cluster_services.pb';
import { usePolicyStyle } from '../../Policies/PolicyStyles';
import { FilterableTable, filterConfig, Link } from '@weaveworks/weave-gitops';
import Severity from '../../Policies/Severity';
import moment from 'moment';
import { TableWrapper } from '../../Shared';
import { useListPolicyValidations } from '../../../contexts/PolicyViolations';
import { Alert } from '@material-ui/lab';
import { LoadingPage } from '@weaveworks/weave-gitops';
import { Field } from '@weaveworks/weave-gitops/ui/components/DataTable';

export enum FieldsType {
  policy = 'POLICY',
  application = 'APPLICATION',
}
interface Props {
  violations: PolicyValidation[];
  tableType?: FieldsType;
  sourcePath?: string;
}

export const PolicyViolationsTable: FC<Props> = ({
  violations,
  tableType = FieldsType.policy,
  sourcePath,
}) => {
  const initialFilterState = {
    ...filterConfig(violations, 'severity'),
  };
  const classes = usePolicyStyle();
  const defaultFields: Field[] = [
    {
      label: 'Severity',
      value: ({ severity }) => <Severity severity={severity || ''} />,
      sortValue: ({ severity }) => severity,
    },
    {
      label: 'Violated Policy',
      value: 'name',
      textSearchable: true,
      sortValue: ({ name }) => name,
    },

    {
      label: 'Violation Time',
      value: (v: PolicyValidation) => moment(v.createdAt).fromNow(),
      defaultSort: true,
      sortValue: ({ createdAt }) => {
        const t = createdAt && new Date(createdAt).getTime();
        return t * -1;
      },
    },
  ];
  const policyFields: Field[] = [
    {
      label: 'Message',
      value: ({ message, clusterName, id }: PolicyValidation) => (
        <Link
          to={`/clusters/violations/details?clusterName=${clusterName}&id=${id}`}
          data-violation-message={message}
        >
          {message}
        </Link>
      ),
      textSearchable: true,
      sortValue: ({ message }) => message,
      maxWidth: 650,
    },
    {
      label: 'Cluster',
      value: 'clusterName',
      sortValue: ({ clusterName }) => clusterName,
    },
    {
      label: 'Application',
      value: (v: PolicyValidation) => `${v.namespace}/${v.entity}`,
    },
    ...defaultFields,
  ];

  const applicationFields: Field[] = [
    {
      label: 'Message',
      value: ({ message, clusterName, id }: PolicyValidation) => (
        <Link
          to={`/clusters/violations/details?clusterName=${clusterName}&id=${id}&source=applications&sourcePath=${sourcePath}`}
          className={classes.link}
          data-violation-message={message}
        >
          {message}
        </Link>
      ),
      textSearchable: true,
      sortValue: ({ message }) => message,
      maxWidth: 650,
    },
    ...defaultFields,
  ];

  const fields =
    tableType === FieldsType.policy ? policyFields : applicationFields;
  return (
    <TableWrapper id="violations-list">
      <FilterableTable
        filters={initialFilterState}
        rows={violations}
        fields={fields}
      ></FilterableTable>
    </TableWrapper>
  );
};

interface PolicyViolationsListProp {
  req: ListPolicyValidationsRequest;
  tableType?: FieldsType;
  sourcePath?: string;
}

export const PolicyViolationsList = ({
  req,
  tableType,
  sourcePath,
}: PolicyViolationsListProp) => {
  const { data, error, isLoading } = useListPolicyValidations(req);

  return (
    <>
      {isLoading && <LoadingPage />}
      {error && <Alert severity="error">{error.message}</Alert>}
      {data?.violations && (
        <PolicyViolationsTable
          violations={data?.violations || []}
          tableType={tableType}
          sourcePath={sourcePath}
        />
      )}
    </>
  );
};
