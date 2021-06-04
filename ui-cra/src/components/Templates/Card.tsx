import React, { FC } from 'react';
import { makeStyles } from '@material-ui/core/styles';
import Card from '@material-ui/core/Card';
import CardActionArea from '@material-ui/core/CardActionArea';
import CardActions from '@material-ui/core/CardActions';
import CardContent from '@material-ui/core/CardContent';
import CardMedia from '@material-ui/core/CardMedia';
import Button from '@material-ui/core/Button';
import Typography from '@material-ui/core/Typography';
import { useHistory } from 'react-router-dom';
import { Template } from '../../types/custom';
import useTemplates from '../../contexts/Templates';

const useStyles = makeStyles({
  root: {
    maxWidth: 345,
  },
});

const TemplateCard: FC<{ template: Template }> = ({ template }) => {
  const classes = useStyles();
  const history = useHistory();
  const { setActiveTemplate } = useTemplates();
  const handleCreateClick = () => {
    setActiveTemplate(template);
    history.push(`/templates/${template.name}/create`);
  };

  return (
    <Card className={classes.root}>
      <CardActionArea>
        <CardMedia
          component="img"
          alt="Stock photo"
          height="140"
          image="https://image.shutterstock.com/image-vector/blue-abstract-striped-textured-geometric-600w-389669770.jpg"
          title="Stock photo"
        />
        <CardContent>
          <Typography gutterBottom variant="h5" component="h2">
            {template.name}
          </Typography>
          <Typography variant="body2" color="textSecondary" component="p">
            {template.description}
          </Typography>
        </CardContent>
      </CardActionArea>
      <CardActions>
        <Button size="medium" color="primary" onClick={handleCreateClick}>
          Create Cluster with this Template
        </Button>
      </CardActions>
    </Card>
  );
};

export default TemplateCard;
