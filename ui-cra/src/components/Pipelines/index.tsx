import { ThemeProvider } from '@material-ui/core';
import styled from 'styled-components';
import { localEEMuiTheme } from '../../muiTheme';
import { useApplicationsCount } from '../Applications/utils';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';

type Props = {
  className?: string;
};

function Pipelines({ className }: Props) {
  const applicationsCount = useApplicationsCount();

  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate documentTitle="WeGo · Pipelines">
        <SectionHeader
          className="count-header"
          path={[
            {
              label: 'Applications',
              url: '/applications/pipelines',
              count: applicationsCount,
            },
            { label: 'Pipelines', count: 0 },
          ]}
        />
        <ContentWrapper loading={false}>
          <div className={className}>
            <p>This is the pipelines page</p>
          </div>
        </ContentWrapper>
      </PageTemplate>
    </ThemeProvider>
  );
}

export default styled(Pipelines).attrs({ className: Pipelines.name })``;
