import React, { FC } from 'react';
import Card from '@material-ui/core/Card';
import CardActionArea from '@material-ui/core/CardActionArea';
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

const Image = styled.div<{ color: string }>`
  background: ${props => props.color};
  height: ${140}px;
`;

const TemplateCard: FC<{ template: Template; color: string }> = ({
  template,
  color,
}) => {
  const history = useHistory();
  const { setActiveTemplate } = useTemplates();
  const handleCreateClick = () => {
    setActiveTemplate(template);
    history.push(`/clusters/templates/${template.name}/create`);
  };

  return (
    <Card data-template-name={template.name}>
      <CardActionArea>
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
        </CardContent>
      </CardActionArea>
      <CardActions>
        <OnClickAction
          id="create-cluster"
          icon={faPlus}
          onClick={handleCreateClick}
          text="CREATE A CLUSTER"
        />
      </CardActions>
    </Card>
  );
};

export default TemplateCard;
