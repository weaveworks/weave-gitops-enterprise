import { FC } from 'react';
import { PageTemplate } from '../Layout/PageTemplate';
import { ContentWrapper } from '../Layout/ContentWrapper';
import {
  LoadingPage,
  theme,
  DataTable,
  filterConfig,
  KubeStatusIndicator,
  filterByStatusCallback,
  statusSortHelper,
  Timestamp,
  formatURL,
} from '@weaveworks/weave-gitops';
import { Link } from 'react-router-dom';
import { makeStyles, createStyles } from '@material-ui/core';
import useGitOpsSets from '../../hooks/gitopssets';
import { Field } from '@weaveworks/weave-gitops/ui/components/DataTable';
import { GitOpsSet } from '../../api/gitopssets/types.pb';
import { computeMessage } from '../Clusters';
import _ from 'lodash';
import { Routes } from '../../utils/nav';

const useStyles = makeStyles(() =>
  createStyles({
    externalIcon: {
      marginRight: theme.spacing.small,
    },
  }),
);

const GitopsSets: FC = () => {
  const { data, isLoading } = useGitOpsSets();
  const gitopssets = data?.gitopssets;

  let initialFilterState = {
    ...filterConfig(gitopssets, 'status', filterByStatusCallback),
    ...filterConfig(gitopssets, 'type'),
    ...filterConfig(gitopssets, 'namespace'),
    ...filterConfig(gitopssets, 'tenant'),
    ...filterConfig(gitopssets, 'clusterName'),
  };

  let fields: Field[] = [
    {
      label: 'Name',
      value: ({ name, namespace }: GitOpsSet) => (
        <Link
          to={formatURL(Routes.GitOpsSetDetail, {
            name,
            namespace,
          })}
        >
          {name}
        </Link>
      ),
      sortValue: ({ name }) => name,
      textSearchable: true,
      maxWidth: 600,
    },
    {
      label: 'Kind',
      value: 'type',
    },
    {
      label: 'Namespace',
      value: 'namespace',
    },
    { label: 'Tenant', value: 'tenant' },
    { label: 'Cluster', value: 'clusterName' },
    {
      label: 'Status',
      value: (gs: GitOpsSet) =>
        gs?.conditions && gs?.conditions?.length > 0 ? (
          <KubeStatusIndicator
            short
            conditions={gs.conditions}
            suspended={false}
          />
        ) : null,
      sortValue: statusSortHelper,
      defaultSort: true,
    },
    {
      label: 'Message',
      value: (gs: GitOpsSet) =>
        (gs?.conditions && computeMessage(gs?.conditions)) || '',
      sortValue: ({ conditions }) => computeMessage(conditions),
      maxWidth: 600,
    },
    {
      label: 'Revision',
      maxWidth: 36,
      value: 'lastAppliedRevision',
    },
    {
      label: 'Last Updated',
      value: (gs: GitOpsSet) => (
        <Timestamp
          time={
            _.get(_.find(gs?.conditions, { type: 'Ready' }), 'timestamp') || ''
          }
        />
      ),
      sortValue: (gs: GitOpsSet) => {
        return _.get(_.find(gs.conditions, { type: 'Ready' }), 'timestamp');
      },
    },
  ];

  return (
    <PageTemplate
      documentTitle="GitopsSets"
      path={[
        {
          label: 'GitopsSets',
        },
      ]}
    >
      <ContentWrapper errors={data?.errors}>
        {isLoading ? (
          <LoadingPage />
        ) : (
          <DataTable
            fields={fields}
            rows={data?.gitopssets}
            filters={initialFilterState}
            hasCheckboxes
          />
        )}
      </ContentWrapper>
    </PageTemplate>
  );
};

export default GitopsSets;
