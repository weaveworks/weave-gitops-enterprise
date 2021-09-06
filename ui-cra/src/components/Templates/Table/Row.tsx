import React, { useCallback } from 'react';
import { TableCell, TableRow, Theme } from '@material-ui/core';
import { createStyles, makeStyles } from '@material-ui/styles';
import { Template } from '../../../types/custom';
import { OnClickAction } from '../../Action';
import { faPlus } from '@fortawesome/free-solid-svg-icons';
import useTemplates from '../../../contexts/Templates';
import { useHistory } from 'react-router-dom';

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    actionButton: {
      fontSize: theme.typography.fontSize,
      margin: `${theme.spacing(0.5)}px ${theme.spacing(1)}px`,
    },
    icon: {
      color: '#ccc',
    },
    nameHeaderCell: {
      paddingLeft: theme.spacing(4),
    },
    nameCell: {
      paddingLeft: theme.spacing(0.5),
    },
    commitsOverviewCell: {
      width: 270,
      padding: 0,
    },
    iconTableCell: {
      width: 30,
    },
    noMaxWidth: {
      maxWidth: 'none',
    },
    normalRow: {
      borderBottom: '1px solid #d8d8d8',
    },
    collapsibleRow: {
      '& > *': {
        paddingTop: 0,
        paddingBottom: 0,
      },
    },
  }),
);

interface RowProps {
  index: number;
  template: Template;
}

const TemplateRow = ({ index, template }: RowProps) => {
  const classes = useStyles();
  const { name, version, description, error } = template;
  const { activeTemplate, setActiveTemplate } = useTemplates();
  const history = useHistory();

  const handleAddCluster = useCallback(() => {
    setActiveTemplate(template);
    history.push(`/clusters/templates/${template.name}/create`);
  }, [setActiveTemplate, history, template]);

  return (
    <TableRow
      className={`summary ${classes.normalRow}`}
      data-template-name={name}
      key={name}
    >
      <TableCell className={classes.nameCell} align="left">
        {template.name}
      </TableCell>
      <TableCell align="left">{description}</TableCell>
      <TableCell>
        <OnClickAction
          id="create-cluster"
          icon={faPlus}
          onClick={handleAddCluster}
          text="CREATE CLUSTER WITH THIS TEMPLATE"
          disabled={Boolean(template.error)}
        />
      </TableCell>
    </TableRow>
  );
};

export default TemplateRow;
