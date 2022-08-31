import { FC } from 'react';
import { PageTemplate } from '../Layout/PageTemplate';
import { SectionHeader } from '../Layout/SectionHeader';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { useApplicationsCount } from './utils';
import {
  AutomationsTable,
  Button,
  Icon,
  IconType,
  LoadingPage,
  useListAutomations,
  theme,
} from '@weaveworks/weave-gitops';
import styled from 'styled-components';
import { Link, useHistory } from 'react-router-dom';
import { useListConfig } from '../../hooks/versions';
import { makeStyles, createStyles } from '@material-ui/core';

interface Size {
  size?: 'small';
}
const ActionsWrapper = styled.div<Size>`
  display: flex;
  & > * {
    margin-right: ${({ theme }) => theme.spacing.small} !important;
  }
`;

const useStyles = makeStyles(() =>
  createStyles({
    externalIcon: {
      marginRight: theme.spacing.small,
    },
  }),
);

const WGApplicationsDashboard: FC = () => {
  const { data: automations, isLoading } = useListAutomations();
  const applicationsCount = useApplicationsCount();
  const history = useHistory();
  const { repoLink } = useListConfig();
  const classes = useStyles();

  const handleAddApplication = () => {
    history.push('/applications/create');
  };

  return (
    <PageTemplate documentTitle="WeGO · Applications">
      <SectionHeader
        path={[
          {
            label: 'Applications',
            url: '/applications',
            count: applicationsCount,
          },
        ]}
      />
      <ContentWrapper errors={automations?.errors}>
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
          }}
        >
          <ActionsWrapper>
            <Button
              id="add-application"
              startIcon={<Icon type={IconType.AddIcon} size="base" />}
              onClick={handleAddApplication}
            >
              ADD AN APPLICATION
            </Button>
            <Link
              target={'_blank'}
              rel="noopener noreferrer"
              component={Button}
              to={{ pathname: repoLink }}
            >
              <Icon
                className={classes.externalIcon}
                type={IconType.ExternalTab}
                size="base"
              />
              GO TO OPEN PULL REQUESTS
            </Link>
          </ActionsWrapper>
        </div>
        {isLoading ? (
          <LoadingPage />
        ) : (
          <AutomationsTable automations={automations?.result} />
        )}
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsDashboard;
