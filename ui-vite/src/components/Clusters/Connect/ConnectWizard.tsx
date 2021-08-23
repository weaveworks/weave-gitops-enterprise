import React, { FC, useState } from 'react';
import styled from 'styled-components';
import theme from 'weaveworks-ui-components/lib/theme';
import { Button } from 'weaveworks-ui-components';

import { ConnectClusterGeneralForm } from './ConnectForm';
import { ConnectClusterConnectionInstructions } from './ConnectionInstructions';
import { ClusterDisconnectionInstructions } from './DisconnectionInstructions';

import { FormState, SetFormState } from '../../../types/form';
import { Cluster } from '../../../types/kubernetes';
import { request } from '../../../utils/request';
import { FlexSpacer } from '../../ListView';
import { HandleFinish } from '../../Shared';

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
  border-bottom: 1px solid ${theme.colors.gray200};
`;

const Title = styled.div<{
  locked: boolean;
  active: boolean;
  onClick?: (ev: Event) => void;
}>`
  color: ${props => (props.locked ? theme.colors.gray200 : theme.colors.black)};
  border-bottom: ${props =>
    props.active
      ? `2px solid ${theme.colors.blue400}`
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
  border-top: 1px solid ${theme.colors.gray200};
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
      onFinish: HandleFinish,
      connecting: boolean,
    ) => (
      <ConnectClusterGeneralForm
        connecting={connecting}
        formState={formState}
        setFormState={setFormState}
      />
    ),
  },
  connect: {
    title: 'Connection instructions',
    content: (formState: FormState, setFormState: SetFormState) => (
      <ConnectClusterConnectionInstructions
        formState={formState}
        setFormState={setFormState}
      />
    ),
  },
  disconnect: {
    title: 'Disconnect',
    content: (
      formState: FormState,
      setFormState: SetFormState,
      onFinish: HandleFinish,
    ) => (
      <ClusterDisconnectionInstructions
        formState={formState}
        setFormState={setFormState}
      />
    ),
  },
};

interface CreateModelProps {
  cluster: Cluster;
  connecting: boolean;
  onFinish: HandleFinish;
}

// SQLite errors
const FRIENDLY_ERRORS: { [key: string]: string } = {
  'UNIQUE constraint failed: clusters.name':
    'Cluster name is already in use and must be unique!',
  'UNIQUE constraint failed: clusters.token':
    'Oops! The token we generated is already in use, this is quite rare. Please try again.',
};

export const ConnectClusterWizard: FC<CreateModelProps> = ({
  connecting,
  cluster,
  onFinish,
}) => {
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

  const onSubmit = (ev: React.FormEvent<HTMLFormElement>) => {
    ev.preventDefault();
    if (!hasNext(formState) || !isValid || submitting) {
      return;
    }
    setFormState({ ...formState, error: '' });
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
      })
      .catch(({ message }) => {
        setFormState({
          ...formState,
          error: FRIENDLY_ERRORS[message] || message,
        });
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
          {content(formState, setFormState, onFinish, connecting)}
        </ContentContainer>
        <ButtonBar>
          <FlexSpacer />
          {formState.activeIndex === 0 && (
            <Button type="submit" disabled={!isValid || submitting}>
              <ButtonText>Save & next</ButtonText>{' '}
              <i className="fas fa-chevron-right" />
            </Button>
          )}
          {formState.activeIndex > 0 && (
            <Button
              className="close-button"
              onClick={() => {
                onFinish({ success: true, message: '' });
              }}
            >
              Close
            </Button>
          )}
        </ButtonBar>
      </form>
    </CreateModelForm>
  );
};
