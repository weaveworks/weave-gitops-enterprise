import { createStyles, makeStyles } from '@material-ui/core';
import Card from '@material-ui/core/Card';
import CardActions from '@material-ui/core/CardActions';
import CardContent from '@material-ui/core/CardContent';
import CardMedia from '@material-ui/core/CardMedia';
import Typography from '@material-ui/core/Typography';
import { Button, Icon, IconType } from '@weaveworks/weave-gitops';
import { FC } from 'react';
import { useHistory } from 'react-router-dom';
import { Template } from '../../cluster-services/cluster_services.pb';
import { ReactComponent as Azure } from '.././assets/img/templates/azure.svg';
import { ReactComponent as DigitalOcean } from '.././assets/img/templates/digitalocean.svg';
import { ReactComponent as Docker } from '.././assets/img/templates/docker.svg';
import { ReactComponent as EKS } from '.././assets/img/templates/eks.svg';
import { ReactComponent as Generic } from '.././assets/img/templates/generic.svg';
import { ReactComponent as GKE } from '.././assets/img/templates/gke.svg';
import { ReactComponent as OpenStack } from '.././assets/img/templates/openstack.svg';
import { ReactComponent as Packet } from '.././assets/img/templates/packet.svg';
import { ReactComponent as VSphere } from '.././assets/img/templates/vsphere.svg';

const useStyles = makeStyles(() =>
  createStyles({
    root: {
      borderRadius: '8px',
      height: '100%',
      display: 'flex',
      flexDirection: 'column',
      justifyContent: 'space-between',
    },
  }),
);

const TemplateCard: FC<{ template: Template }> = ({ template }) => {
  const classes = useStyles();
  const history = useHistory();
  const handleCreateClick = () => {
    history.push(`/templates/${template.name}/create`);
  };
  const getTile = () => {
    switch (template.provider) {
      case 'azure':
        return <Azure />;
      case 'aws':
        return <EKS />;
      case 'digitalocean':
        return <DigitalOcean />;
      case 'docker':
        return <Docker />;
      case 'gcp':
        return <GKE />;
      case 'openstack':
        return <OpenStack />;
      case 'packet':
        return <Packet />;
      case 'vsphere':
        return <VSphere />;
      default:
        return <Generic />;
    }
  };

  const disabled = Boolean(template.error);

  return (
    <Card className={classes.root} data-template-name={template.name}>
      <CardMedia>{getTile()}</CardMedia>
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
        <Button
          id="create-cluster"
          startIcon={<Icon type={IconType.AddIcon} size="base" />}
          onClick={handleCreateClick}
          disabled={disabled}
        >
          CREATE A CLUSTER
        </Button>
      </CardActions>
    </Card>
  );
};

export default TemplateCard;
