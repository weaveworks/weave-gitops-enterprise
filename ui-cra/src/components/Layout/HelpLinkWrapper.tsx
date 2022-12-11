import { Link, theme } from '@weaveworks/weave-gitops';

import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { Tooltip } from '../Shared';
import styled from 'styled-components';

import { useListConfig, useListVersion } from '../../hooks/versions';
import { makeStyles, createStyles } from '@material-ui/core/styles';
import { toast } from 'react-toastify';

const { xxs, xs, medium } = theme.spacing;

const HelpLink = styled.div<{
  backgroundColor?: string;
  textColor?: string;
}>`
  padding: calc(${medium} - ${xxs}) ${medium};
  background-color: ${props =>
    props.backgroundColor || 'rgba(255, 255, 255, 0.7)'};
  color: ${props => props.textColor || theme.colors.neutral30};
  border-radius: 0 0 ${xs} ${xs};
  display: flex;
  justify-content: space-between;
  a {
    color: ${({ theme }) => theme.colors.primary};
  }
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
const HelpLinkWrapper = () => {
  const { data, error } = useListVersion();
  const { uiConfig } = useListConfig();
  const versions = {
    capiServer: data?.data.version,
    ui: process.env.REACT_APP_VERSION || 'no version specified',
  };
  const classes = useStyles();

  if (error) {
    toast.error(error.message);
  }
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
          <Link newTab href="https://weavesupport.zendesk.com/">
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

export default HelpLinkWrapper;
