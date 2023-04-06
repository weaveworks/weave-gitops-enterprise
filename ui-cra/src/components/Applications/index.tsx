import { FC } from 'react';
import { PageTemplate } from '../Layout/PageTemplate';
import { ContentWrapper } from '../Layout/ContentWrapper';
import {
  AutomationsTable,
  Button,
  Icon,
  IconType,
  LoadingPage,
  useListAutomations,
  theme,
  Flex,
} from '@weaveworks/weave-gitops';
import styled from 'styled-components';
import { useHistory } from 'react-router-dom';
import { makeStyles, createStyles } from '@material-ui/core';
import { openLinkHandler } from '../../utils/link-checker';
import { Routes } from '../../utils/nav';
import { useListConfigContext } from '../../contexts/ListConfig';

interface Size {
  size?: 'small';
}
const ActionsWrapper = styled(Flex)<Size>`
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
  const { data: automations, isLoading } = useListAutomations();
  const history = useHistory();
  const listConfigContext = useListConfigContext();
  const repoLink = listConfigContext?.repoLink || '';
  const classes = useStyles();

  const handleAddApplication = () => history.push(Routes.AddApplication);

  return (
    <PageTemplate
      documentTitle="Applications"
      path={[
        {
          label: 'Applications',
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
          <AutomationsTable automations={automations?.result} />
        )}
      </ContentWrapper>
    </PageTemplate>
  );
};

export default WGApplicationsDashboard;
