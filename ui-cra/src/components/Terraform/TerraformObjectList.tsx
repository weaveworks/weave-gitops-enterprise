import { ThemeProvider } from '@material-ui/core';
import {
  DataTable,
  formatURL,
  KubeStatusIndicator,
} from '@weaveworks/weave-gitops';

import { Link } from 'react-router-dom';
import styled from 'styled-components';
import { TerraformObject } from '../../api/terraform/types.pb';
import { useListTerraformObjects } from '../../contexts/Terraform';
import { localEEMuiTheme } from '../../muiTheme';
import { computeMessage } from '../../utils/conditions';
import { getKindRoute, Routes } from '../../utils/nav';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { TableWrapper } from '../Shared';
type Props = {
  className?: string;
};

function TerraformObjectList({ className }: Props) {
  const { isLoading, data, error } = useListTerraformObjects();

  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate documentTitle="WeGo · Terraform">
        <SectionHeader
          className="count-header"
          path={[
            {
              label: 'Terraform Objects',
              url: '/terraform',
              count: data?.objects?.length,
            },
          ]}
        />

        <ContentWrapper
          errors={error ? [error] : data?.errors || []}
          loading={isLoading}
        >
          <TableWrapper>
            <DataTable
              className={className}
              fields={[
                {
                  value: ({
                    name,
                    namespace,
                    clusterName,
                  }: TerraformObject) => (
                    <Link
                      to={formatURL(Routes.TerraformDetail, {
                        name,
                        namespace,
                        clusterName,
                      })}
                    >
                      {name}
                    </Link>
                  ),
                  label: 'Name',
                },
                { value: 'namespace', label: 'Namespace' },
                { value: () => '-', label: 'Tenant' },
                { value: 'clusterName', label: 'Cluster' },
                {
                  label: 'Source',
                  value: (tf: TerraformObject) => {
                    const route = getKindRoute(tf.sourceRef?.kind as string);

                    const { name, namespace } = tf.sourceRef || {};

                    const u = formatURL(route, {
                      clusterName: tf.clusterName,
                      name,
                      namespace,
                    });

                    return <Link to={u}>{name}</Link>;
                  },
                },
                {
                  value: (tf: TerraformObject) => (
                    <KubeStatusIndicator conditions={tf.conditions || []} />
                  ),
                  label: 'Status',
                },
                {
                  value: (tf: TerraformObject) =>
                    computeMessage(tf.conditions as any),
                  label: 'Message',
                },
              ]}
              rows={data?.objects}
            />
          </TableWrapper>
        </ContentWrapper>
      </PageTemplate>
    </ThemeProvider>
  );
}

export default styled(TerraformObjectList).attrs({
  className: TerraformObjectList.name,
})``;
