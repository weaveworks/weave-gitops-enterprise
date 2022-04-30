import React, { FC, useState } from 'react';
import styled from 'styled-components';
import { Button, theme } from '@weaveworks/weave-gitops';
import { FormState, SetFormState } from '../../../types/form';
import useNotifications from './../../../contexts/Notifications';
import { GitopsCluster } from '../../../capi-server/capi_server.pb';

export const ButtonText = styled.span`
  margin: 0 ${theme.spacing.xxs};
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

// SQLite errors
const FRIENDLY_ERRORS: { [key: string]: string } = {
  'UNIQUE constraint failed: clusters.name':
    'Cluster name is already in use and must be unique!',
  'UNIQUE constraint failed: clusters.token':
    'Oops! The token we generated is already in use, this is quite rare. Please try again.',
};

export const ConnectClusterWizard: FC<{
  cluster: GitopsCluster;
  connecting: boolean;
  onFinish: () => void;
}> = ({ connecting, cluster, onFinish }) => {
  return (
    <CreateModelForm>
      <form>
        {/* <ContentContainer>
          {content(formState, setFormState, connecting, onFinish)}
        </ContentContainer> */}
        <ButtonBar>
          <div style={{ flex: 1 }} />
          <Button
            type="submit"
            startIcon={<i className="fas fa-chevron-right" />}
          >
            SAVE & NEXT
          </Button>
        </ButtonBar>
      </form>
    </CreateModelForm>
  );
};
