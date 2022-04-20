import moment from 'moment';
import { PolicyValidation } from '../../../capi-server/capi_server.pb';
import { PolicyStyles } from '../../Policies/PolicyStyles';
import Severity from '../../Policies/Severity';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { darcula } from 'react-syntax-highlighter/dist/esm/styles/prism';
import MDEditor from '@uiw/react-md-editor';

interface IViolationDetailsProps {
  violation: PolicyValidation | undefined;
}

function ViolationDetails({ violation }: IViolationDetailsProps) {
  const classes = PolicyStyles.useStyles();
  const {
    severity,
    createdAt,
    category,
    howToSolve,
    message,
    description,
    violatingEntity,
  } = violation || {};

  return (
    <>
      <div className={`${classes.contentWrapper} ${classes.flexStart}`}>
        <div className={classes.cardTitle}>Message:</div>
        <span className={classes.body1}>{message}</span>
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

      <hr />
      <div className={classes.sectionSeperator}>
        <div className={classes.cardTitle}>Description:</div>
        <MDEditor.Markdown source={description} className={classes.editor} />
      </div>

      <div className={classes.sectionSeperator}>
        <div className={classes.cardTitle}>How to solve:</div>
        <MDEditor.Markdown source={howToSolve} className={classes.editor} />
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
            {violatingEntity}
          </SyntaxHighlighter>
        </div>
      </div>
    </>
  );
}

export default ViolationDetails;
