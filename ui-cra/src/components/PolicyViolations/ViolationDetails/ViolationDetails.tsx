import moment from 'moment';
import { PolicyValidation } from '../../../cluster-services/cluster_services.pb';
import { usePolicyStyle } from '../../Policies/PolicyStyles';
import Severity from '../../Policies/Severity';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { darcula } from 'react-syntax-highlighter/dist/esm/styles/prism';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';

interface IViolationDetailsProps {
  violation: PolicyValidation | undefined;
}

function ViolationDetails({ violation }: IViolationDetailsProps) {
  const classes = usePolicyStyle();
  const {
    severity,
    createdAt,
    category,
    howToSolve,
    message,
    description,
    violatingEntity,
    entity,
    namespace,
  } = violation || {};

  return (
    <>
      <div className={`${classes.contentWrapper} ${classes.flexStart}`}>
        <div className={classes.cardTitle}>Message:</div>
        <span className={classes.body1}>{message}</span>
      </div>
      <div className={`${classes.contentWrapper} ${classes.flexStart}`}>
        <div className={classes.cardTitle}>Cluster Name:</div>
        <span className={classes.body1}>{clusterName}</span>
      </div>

      <div className={`${classes.contentWrapper} ${classes.flexStart}`}>
        <div className={classes.cardTitle}>Violation Time:</div>
        <span className={classes.body1}>{moment(createdAt).fromNow()}</span>
      </div>
      <div className={`${classes.contentWrapper} ${classes.flexStart}`}>
        <div className={`${classes.cardTitle} ${classes.marginrightSmall}`}>
          Severity:
        </div>
        <Severity severity={severity || ''} />
      </div>
      <div className={`${classes.contentWrapper} ${classes.flexStart}`}>
        <div className={classes.cardTitle}>Category:</div>
        <span className={classes.body1}>{category}</span>
      </div>
      <div className={`${classes.contentWrapper} ${classes.flexStart}`}>
        <div className={classes.cardTitle}>Application:</div>
        <span className={classes.body1}>{`${namespace}/${entity}`}</span>
      </div>
      <hr />
      <div className={classes.sectionSeperator}>
        <div className={classes.cardTitle}>Description:</div>
        <ReactMarkdown
          children={description || ''}
          className={classes.editor}
        />
      </div>

      <div className={classes.sectionSeperator}>
        <div className={classes.cardTitle}>How to solve:</div>
        <ReactMarkdown
          children={howToSolve || ''}
          className={classes.editor}
          remarkPlugins={[remarkGfm]}
        />{' '}
      </div>

      <div className={classes.sectionSeperator}>
        <div className={classes.cardTitle}>Violating Entity:</div>
        <div>
          <SyntaxHighlighter
            language="json"
            style={darcula}
            wrapLongLines="pre-wrap"
            showLineNumbers={true}
            codeTagProps={{
              className: classes.code,
            }}
            customStyle={{
              height: '450px',
            }}
          >
            {JSON.parse(JSON.stringify(violatingEntity, null, 2))}
          </SyntaxHighlighter>
        </div>
      </div>
    </>
  );
}

export default ViolationDetails;
