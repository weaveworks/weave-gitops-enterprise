import { ThemeProvider } from '@material-ui/core';
import styled from 'styled-components';
import { localEEMuiTheme } from '../../muiTheme';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';

type Props = {
  className?: string;
};

function ObjectList({ className }: Props) {
  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate documentTitle="WeGo · Terraform">
        <SectionHeader
          className="count-header"
          path={[{ label: 'Terraform Objects', url: '/terraform' }]}
        />

        <ContentWrapper>tf tingz</ContentWrapper>
      </PageTemplate>
      <div className={className}></div>
    </ThemeProvider>
  );
}

export default styled(ObjectList).attrs({ className: ObjectList.name })``;
