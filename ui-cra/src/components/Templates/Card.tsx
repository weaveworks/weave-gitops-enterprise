import React, { FC } from 'react';
import Card from '@material-ui/core/Card';
import CardActions from '@material-ui/core/CardActions';
import CardContent from '@material-ui/core/CardContent';
import CardMedia from '@material-ui/core/CardMedia';
import Typography from '@material-ui/core/Typography';
import { useHistory } from 'react-router-dom';
import { Template } from '../../types/custom';
import useTemplates from '../../contexts/Templates';
import styled from 'styled-components';
import { OnClickAction } from '../Action';
import { faPlus } from '@fortawesome/free-solid-svg-icons';
import { createStyles, makeStyles } from '@material-ui/core';

const Image = styled.div<{ color: string }>`
  background: ${props => props.color};
  height: ${140}px;
`;

const useStyles = makeStyles(() =>
  createStyles({
    root: {
      borderRadius: '8px',
    },
  }),
);

const TemplateCard: FC<{ template: Template; color: string }> = ({
  template,
  color,
}) => {
  const classes = useStyles();
  const history = useHistory();
  const { setActiveTemplate } = useTemplates();
  const handleCreateClick = () => {
    setActiveTemplate(template);
    history.push(`/clusters/templates/${template.name}/create`);
  };

  const disabled = Boolean(template.error);
  console.log({ disabled });

  return (
    <Card className={classes.root} data-template-name={template.name}>
      <CardMedia>
        <Image color={color} />
      </CardMedia>
      <CardContent>
        <Typography gutterBottom variant="h6" component="h2">
          {template.name}
        </Typography>
        <Typography variant="body2" color="textSecondary" component="p">
          {template.description}
        </Typography>
        {template.error && (
          <>
            <Typography
              className="template-error-header"
              variant="h6"
              component="h2"
              color="error"
            >
              Error in template
            </Typography>
            <Typography
              className="template-error-description"
              variant="body2"
              color="error"
              component="p"
            >
              {template.error}
            </Typography>
          </>
        )}
      </CardContent>
      <CardActions>
        <OnClickAction
          id="create-cluster"
          disabled={disabled}
          icon={faPlus}
          onClick={handleCreateClick}
          text="CREATE A CLUSTER"
        />
      </CardActions>
    </Card>
  );
};

export default TemplateCard;
