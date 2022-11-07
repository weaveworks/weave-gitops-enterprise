import { theme } from '@weaveworks/weave-gitops';
import styled from 'styled-components';

const Message = styled.div`
  background: rgba(255, 255, 255, 0.85);
  box-shadow: 5px 10px 50px 3px rgb(0 0 0 / 10%);
  border-radius: 10px;
  padding: ${theme.spacing.large} ${theme.spacing.xxl};
  max-width: 560px;
  margin: auto;
  display: flex;
  flex-direction: column;
`;
const Title = styled.h4`
  font-size: ${theme.fontSizes.large};
  font-weight: 600;
  color: ${theme.colors.neutral30};
  margin-bottom: ${theme.spacing.small};
`;

const Body = styled.p`
  font-size: ${theme.fontSizes.medium};
  color: ${theme.colors.neutral30};
  font-weight: 400;
`;

const NoRunsMessage = () => {
  return (
    <Message>
      <Title>GitOps Run</Title>
      <Body>
        There are no instances of GitOps Run running on the cluster at the
        moment
      </Body>
    </Message>
  );
};

export default NoRunsMessage;
