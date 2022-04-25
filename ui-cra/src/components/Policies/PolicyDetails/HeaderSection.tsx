import { Policy } from '../../../capi-server/capi_server.pb';
import Severity from '../Severity';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { darcula } from 'react-syntax-highlighter/dist/esm/styles/prism';
import { usePolicyStyle } from '../PolicyStyles';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm'

function HeaderSection({
  id,
  tags,
  severity,
  category,
  targets,
  description,
  howToSolve,
  code,
}: Policy) {
  const classes = usePolicyStyle();

  return (
    <>
      <div className={`${classes.contentWrapper} ${classes.flexStart}`}>
        <div className={classes.cardTitle}>Policy ID:</div>
        <span className={classes.body1}>{id}</span>
      </div>
      <div className={`${classes.contentWrapper} ${classes.flexStart}`}>
        <span className={classes.cardTitle}>Tags:</span>
        {!!tags && tags?.length > 0 ? (
          tags?.map(tag => (
            <span key={tag} className={classes.chip}>
              {tag}
            </span>
          ))
        ) : (
          <span className={classes.body1}>
            There is no tags for this policy
          </span>
        )}
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
        <div className={classes.cardTitle}>Targeted K8s Kind:</div>
        {targets?.kinds?.map(kind => (
          <span key={kind} className={classes.chip}>
            {kind}
          </span>
        ))}
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
        <ReactMarkdown children={howToSolve || ''} className={classes.editor} remarkPlugins={[remarkGfm]}  />
      </div>

      <div className={classes.sectionSeperator}>
        <div className={classes.cardTitle}>Policy Code:</div>
        <div>
          <SyntaxHighlighter
            language="rego"
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
            {code}
          </SyntaxHighlighter>
        </div>
      </div>
    </>
  );
}

export default HeaderSection;
