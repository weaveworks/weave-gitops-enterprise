import moment from 'moment';
import { PolicyValidation } from '../../../cluster-services/cluster_services.pb';
import { usePolicyStyle } from '../../Policies/PolicyStyles';
import Severity from '../../Policies/Severity';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { darcula } from 'react-syntax-highlighter/dist/esm/styles/prism';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { Link } from 'react-router-dom';
import { generateRowHeaders } from '../../RowHeader';
import PolicyConfigSection from './PolicyConfigSection';

interface IViolationDetailsProps {
  violation: PolicyValidation | undefined;
  source?: string;
}

function ViolationDetails({ violation, source }: IViolationDetailsProps) {
  const classes = usePolicyStyle();
  const {
    severity,
    createdAt,
    category,
    howToSolve,
    description,
    violatingEntity,
    entity,
    namespace,
    occurrences,
    clusterName,
    name,
    parameters,
  } = violation || {};
  const defaultHeaders = [
    {
      rowkey: 'Cluster',
      value: clusterName,
    },
    {
      rowkey: 'Violation Time',
      value: moment(createdAt).fromNow(),
    },
    {
      rowkey: 'Severity',
      children: <Severity severity={severity || ''} />,
    },
    {
      rowkey: 'Category',
      value: category,
    },
  ];

  const applicationHeaderDetails = [
    {
      rowkey: 'Policy Name',
      children: (
        <Link
          to={`/policies/details?clusterName=${clusterName}&id=${violation?.policyId}`}
          className={classes.link}
          data-violation-message={name}
        >
          {name}
        </Link>
      ),
    },
    ...defaultHeaders,
  ];
  const headerDetails = [
    ...defaultHeaders,
    {
      rowkey: 'Application',
      value: `${namespace}/${entity}`,
    },
  ];
  const displayedHeaders = !!source ? applicationHeaderDetails : headerDetails;

  return (
    <>
      {generateRowHeaders(displayedHeaders)}

      <div className={classes.sectionSeperator}>
        <div className={classes.cardTitle}>
          Occurrences{' '}
          <span className={classes.titleNotification}>
            ( {occurrences?.length} )
          </span>
        </div>
        <ul className={classes.occurrencesList} id="occurrences">
          {occurrences?.map(item => (
            <li key={item.message} className={classes.body1}>
              {item.message}
            </li>
          ))}
        </ul>
      </div>

      <div className={classes.sectionSeperator} data-testid="description">
        <div className={classes.cardTitle}>Description:</div>
        <ReactMarkdown
          children={description || ''}
          className={classes.editor}
        />
      </div>

      <div className={classes.sectionSeperator} data-testid="howToSolve">
        <div className={classes.cardTitle}>How to solve:</div>
        <ReactMarkdown
          children={howToSolve || ''}
          className={classes.editor}
          remarkPlugins={[remarkGfm]}
        />
      </div>

      <div className={classes.sectionSeperator} data-testid="violatingEntity">
        <div className={classes.cardTitle}>Violating Entity:</div>
        <div>
          <SyntaxHighlighter
            language="json"
            style={darcula}
            wrapLongLines
            showLineNumbers
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

      <div className={classes.sectionSeperator}>
        <PolicyConfigSection parameters={parameters} />
      </div>
    </>
  );
}

export default ViolationDetails;
