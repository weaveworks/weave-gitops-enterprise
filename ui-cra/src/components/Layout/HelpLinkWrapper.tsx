import { Flex, Link } from '@weaveworks/weave-gitops';

import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import styled from 'styled-components';
import { Tooltip } from '../Shared';

import { createStyles, makeStyles } from '@material-ui/core/styles';
import React from 'react';
import {
  useListConfigContext,
  useVersionContext,
} from '../../contexts/ListConfig';

const HelpLink = styled(Flex)<{
  backgroundColor?: string;
  textColor?: string;
}>`
  padding: calc(
      ${props => props.theme.spacing.medium} -
        ${props => props.theme.spacing.xxs}
    )
    ${props => props.theme.spacing.medium};
  background-color: ${props =>
    props.backgroundColor || props.theme.colors.neutralGray};
  color: ${props => props.textColor || props.theme.colors.neutral30};
  border-radius: 0 0 ${props => props.theme.spacing.xs}
    ${props => props.theme.spacing.xs};
  justify-content: space-between;
`;

const useStyles = makeStyles(() =>
  createStyles({
    editor: {
      '& p': {
        margin: 0,
      },
    },
  }),
);
const Footer = ({ version }: { version: string }) => {
  const classes = useStyles();
  const listConfigContext = useListConfigContext();
  const uiConfig = listConfigContext?.uiConfig || '';

  const versions = {
    capiServer: version,
    ui: process.env.REACT_APP_VERSION || 'no version specified',
  };

  return (
    <HelpLink
      backgroundColor={uiConfig?.footer?.backgroundColor}
      textColor={uiConfig?.footer?.color}
    >
      {uiConfig?.footer?.content ? (
        <div>
          <ReactMarkdown
            children={uiConfig?.footer?.content || ''}
            remarkPlugins={[remarkGfm]}
            className={classes.editor}
          />
        </div>
      ) : (
        <div>
          Need help? Raise a&nbsp;
          <Link newTab href="https://support.weave.works/helpdesk/">
            support ticket
          </Link>
        </div>
      )}
      {!uiConfig?.footer?.hideVersion ? (
        <Tooltip
          title={`Server Version ${versions?.capiServer}`}
          placement="top"
        >
          <div>Weave GitOps Enterprise {process.env.REACT_APP_VERSION}</div>
        </Tooltip>
      ) : null}
    </HelpLink>
  );
};

const HelpLinkWrapper = () => {
  const versionResponse = useVersionContext();
  return <Footer version={versionResponse?.data.version} />;
};
const MemoizedHelpLinkWrapper = React.memo(HelpLinkWrapper);

export default MemoizedHelpLinkWrapper;
