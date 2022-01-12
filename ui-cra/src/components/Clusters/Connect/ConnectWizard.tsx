import React, { FC, useState } from 'react';
import styled from 'styled-components';
import { Button, theme } from '@weaveworks/weave-gitops';
import { ConnectClusterGeneralForm } from './ConnectForm';
import { ConnectClusterConnectionInstructions } from './ConnectionInstructions';
import { ClusterDisconnectionInstructions } from './DisconnectionInstructions';
import { FormState, SetFormState } from '../../../types/form';
import { Cluster } from '../../../types/kubernetes';
import { request } from '../../../utils/request';
import useNotifications from './../../../contexts/Notifications';

export const ButtonText = styled.span`
  margin: 0 4px;
`;

export const ContentContainer = styled.div`
  margin: ${theme.spacing.base} 0;
  min-height: 250px;
  overflow-y: auto;
`;

const TitleBarContainer = styled.div`
  display: flex;
  margin-bottom: ${theme.spacing.base};
  border-bottom: 1px solid ${theme.colors.neutral20};
`;

const Title = styled.div<{
  locked: boolean;
  active: boolean;
  onClick?: (ev: Event) => void;
}>`
  color: ${props =>
    props.locked ? theme.colors.neutral30 : theme.colors.black};
  border-bottom: ${props =>
    props.active
      ? `2px solid ${theme.colors.primary}`
      : '2px solid transparent'};
  padding: ${theme.spacing.small} ${theme.spacing.base};
  cursor: ${props => (props.onClick ? 'pointer' : 'default')};
`;

const CreateModelForm = styled.div``;

interface TitleBarProps {
  activeIndex: number;
  titles: string[];
  locked: boolean;
  onClick: (index: number) => void;
}

const TitleBar: FC<TitleBarProps> = ({
  locked,
  titles,
  activeIndex,
  onClick,
}) => (
  <TitleBarContainer>
    {titles.map((title, i) => (
      <Title
        locked={locked && i !== 0}
        onClick={locked ? undefined : () => onClick(i)}
        active={i === activeIndex}
        key={title}
      >
        {title}
      </Title>
    ))}
  </TitleBarContainer>
);

const ButtonBar = styled.div`
  display: flex;
  border-top: 1px solid ${theme.colors.neutral20};
  padding-top: ${theme.spacing.small};
`;

const nextPage = (formState: FormState): FormState => ({
  ...formState,
  activeIndex: formState.activeIndex + 1,
});
const hasNext = (formState: FormState): boolean =>
  formState.activeIndex < formState.numberOfItems - 1;

const PAGES = {
  general: {
    title: 'General',
    content: (
      formState: FormState,
      setFormState: SetFormState,
      connecting: boolean,
    ) => (
      <ConnectClusterGeneralForm
        formState={formState}
        setFormState={setFormState}
        connecting={connecting}
      />
    ),
  },
  connect: {
    title: 'Connection instructions',
    content: (
      formState: FormState,
      setFormState: SetFormState,
      connecting: boolean,
    ) => (
      <ConnectClusterConnectionInstructions
        formState={formState}
        setFormState={setFormState}
        connecting={connecting}
      />
    ),
  },
  disconnect: {
    title: 'Disconnect',
    content: (
      formState: FormState,
      setFormState: SetFormState,
      connecting: boolean,
      onFinish: () => void,
    ) => (
      <ClusterDisconnectionInstructions
        formState={formState}
        setFormState={setFormState}
        connecting={connecting}
        onFinish={onFinish}
      />
    ),
  },
};

// SQLite errors
const FRIENDLY_ERRORS: { [key: string]: string } = {
  'UNIQUE constraint failed: clusters.name':
    'Cluster name is already in use and must be unique!',
  'UNIQUE constraint failed: clusters.token':
    'Oops! The token we generated is already in use, this is quite rare. Please try again.',
};

export const ConnectClusterWizard: FC<{
  cluster: Cluster;
  connecting: boolean;
  onFinish: () => void;
}> = ({ connecting, cluster, onFinish }) => {
  const pages = connecting
    ? [PAGES.general, PAGES.connect]
    : [PAGES.general, PAGES.connect, PAGES.disconnect];
  const [formState, setFormState] = useState<FormState>({
    activeIndex: 0,
    numberOfItems: pages.length,
    cluster,
    error: '',
  });
  const [submitting, setSubmitting] = useState<boolean>(false);
  const titles = pages.map(page => page.title);
  const { content } = pages[formState.activeIndex];
  const isValid = formState.cluster.name.trim() !== '';
  const { setNotifications } = useNotifications();

  const onSubmit = (ev: React.FormEvent<HTMLFormElement>) => {
    ev.preventDefault();
    if (!hasNext(formState) || !isValid || submitting) {
      return;
    }
    setFormState({ ...formState });
    setSubmitting(true);
    const id = formState.cluster.id;
    const req = id
      ? request('PUT', `/gitops/api/clusters/${id}`, {
          body: JSON.stringify(formState.cluster),
        })
      : request('POST', '/gitops/api/clusters', {
          body: JSON.stringify(formState.cluster),
        });

    req
      .then((cluster: Cluster) => {
        setSubmitting(false);
        setFormState({ ...formState, cluster });
        setFormState(nextPage);
        setNotifications([
          {
            message: 'Cluster successfully added to the MCCP',
            variant: 'success',
          },
        ]);
      })
      .catch(({ message }) => {
        setNotifications([
          {
            message: FRIENDLY_ERRORS[message] || message,
            variant: 'danger',
          },
        ]);
        setSubmitting(false);
      });
  };

  const setActiveIndex = (activeIndex: number) =>
    setFormState({ ...formState, activeIndex });

  return (
    <CreateModelForm>
      <TitleBar
        onClick={index => setActiveIndex(index)}
        locked={!formState.cluster.id}
        activeIndex={formState.activeIndex}
        titles={titles}
      />
      <form onSubmit={onSubmit}>
        <ContentContainer>
          {content(formState, setFormState, connecting, onFinish)}
        </ContentContainer>
        <ButtonBar>
          <div style={{ flex: 1 }} />
          {formState.activeIndex === 0 && (
            <Button
              type="submit"
              startIcon={<i className="fas fa-chevron-right" />}
              disabled={!isValid || submitting}
            >
              SAVE & NEXT
            </Button>
          )}
          {formState.activeIndex > 0 && (
            <Button className="close-button" onClick={() => onFinish()}>
              CLOSE
            </Button>
          )}
        </ButtonBar>
      </form>
    </CreateModelForm>
  );
};
