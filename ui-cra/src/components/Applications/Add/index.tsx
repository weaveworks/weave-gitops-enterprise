import React from 'react';
import { ThemeProvider } from '@material-ui/core/styles';
import { localEEMuiTheme } from '../../../muiTheme';
import { PageTemplate } from '../../Layout/PageTemplate';
import { SectionHeader } from '../../Layout/SectionHeader';
import { useApplicationsCount } from '../utils';

const AddApplication = () => {
  const applicationsCount = useApplicationsCount();

  return (
    <ThemeProvider theme={localEEMuiTheme}>
      <PageTemplate documentTitle="WeGo · Add new application">
        <SectionHeader
          className="count-header"
          path={[
            {
              label: 'Applications',
              url: '/applications',
              count: applicationsCount,
            },
            { label: 'Add new application' },
          ]}
        />
      </PageTemplate>
    </ThemeProvider>
  );
};
export default AddApplication;
