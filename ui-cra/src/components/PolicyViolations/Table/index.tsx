import { FC } from 'react';
import {
  ListPolicyValidationsRequest,
  PolicyValidation,
} from '../../../cluster-services/cluster_services.pb';
import { usePolicyStyle } from '../../Policies/PolicyStyles';
import { FilterableTable, filterConfig } from '@weaveworks/weave-gitops';
import { Link } from 'react-router-dom';
import Severity from '../../Policies/Severity';
import moment from 'moment';
import { TableWrapper } from '../../Shared';
import { useListPolicyValidations } from '../../../contexts/PolicyViolations';
import { Alert } from '@material-ui/lab';
import { LoadingPage } from '@weaveworks/weave-gitops';

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
  sourcePath
}) => {
  const initialFilterState = {
    ...filterConfig(violations, 'severity'),
  };
  const classes = usePolicyStyle();
  const policyFields = [
    {
      label: 'Name configured in management UI',
      value: ({ message, clusterName, id }: PolicyValidation) => (
        <Link
          to={`/clusters/violations/details?clusterName=${clusterName}&id=${id}`}
          className={classes.link}
          data-violation-message={message}
        >
          {message}
        </Link>
      ),
      textSearchable: true,
      maxWidth: 650,
    },
    {
      label: 'Message',
      value: ({ message, clusterName, id }: PolicyValidation) => (
        <Link
          to={`/clusters/violations/details?clusterName=${clusterName}&id=${id}`}
          className={classes.link}
          data-violation-message={message}
        >
          {message}
        </Link>
      ),
      textSearchable: true,
      maxWidth: 650,
    },
    {
      label: 'Severity',
      value: (v: PolicyValidation) => <Severity severity={v.severity || ''} />,
    },
    {
      label: 'Violated Policy',
      value: 'name',
    },
    {
      label: 'Cluster',
      value: 'clusterName',
    },
    {
      label: 'Violation Time',
      value: (v: PolicyValidation) => moment(v.createdAt).fromNow(),
    },
    {
      label: 'Application',
      value: (v: PolicyValidation) => `${v.namespace}/${v.entity}`,
    },
  ];

  const applicationFields = [
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
      maxWidth: 650,
    },
    {
      label: 'Severity',
      value: (v: PolicyValidation) => <Severity severity={v.severity || ''} />,
    },
    {
      label: 'Violated Policy',
      value: 'name',
    },
    {
      label: 'Violation Time',
      value: (v: PolicyValidation) => moment(v.createdAt).fromNow(),
    },
  ];

  const fields =
    tableType === FieldsType.policy ? policyFields : applicationFields;
  return (
    <TableWrapper>
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
  sourcePath
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
