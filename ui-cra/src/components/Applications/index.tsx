import { createStyles, makeStyles } from '@material-ui/core';
import {
  AutomationsTable,
  Button,
  Icon,
  IconType,
  LoadingPage, theme, useListAutomations
} from '@weaveworks/weave-gitops';
import { FC } from 'react';
import { useHistory } from 'react-router-dom';
import styled from 'styled-components';
import { useListConfig } from '../../hooks/versions';
import { openLinkHandler } from '../../utils/link-checker';
import { Routes } from '../../utils/nav';
import { ContentWrapper } from '../Layout/ContentWrapper';
import { PageTemplate } from '../Layout/PageTemplate';

interface Size {
  size?: 'small';
}
const ActionsWrapper = styled.div<Size>`
  display: flex;
  & > .actionButton.btn {
    margin-right: ${({ theme }) => theme.spacing.small};
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
  const { data: automations, isLoading, error } = useListAutomations();
  const history = useHistory();
  const { repoLink } = useListConfig();
  const classes = useStyles();

  const handleAddApplication = () => history.push(Routes.AddApplication);

  return (
    <PageTemplate
      documentTitle="Applications"
      path={[
        {
          label: 'Applications',
          url: Routes.Applications,
        },
      ]}
    >
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
              className="actionButton btn"
              startIcon={<Icon type={IconType.AddIcon} size="base" />}
              onClick={handleAddApplication}
            >
              ADD AN APPLICATION
            </Button>
            <Button onClick={openLinkHandler(repoLink)}>
              <Icon
                className={classes.externalIcon}
                type={IconType.ExternalTab}
                size="base"
              />
              GO TO OPEN PULL REQUESTS
            </Button>
          </ActionsWrapper>
        </div>
        {isLoading ? (
          <LoadingPage />
        ) : (
          automations && <AutomationsTable automations={automations?.result} />
          
        )}
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsDashboard;
