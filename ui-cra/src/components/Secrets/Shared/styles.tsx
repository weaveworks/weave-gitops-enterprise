import styled from 'styled-components';

export const FormWrapper = styled.form`
  .group-section {
    width: 100%;
    .form-group {
      display: flex;
      flex-direction: column;
    }
    .form-section {
      width: 40%;
    }
    .MuiRadio-colorSecondary.Mui-checked {
      color: ${props => props.theme.colors.primary10};
    }
    h2 {
      font-size: 20px;
      margin-bottom: ${props => props.theme.spacing.xs};
    }
  }
  .MuiInputBase-input {
    padding-left: 8px;
  }
  .form-section {
    width: 40%;
    margin-right: 24px;
  }
  .auth-message {
    padding-right: 0;
    margin-left: 24px;
  }
  .secret-data-list {
    display: flex;
    align-items: self-start;
    width: 100%;
    .remove-icon {
      margin-top: 25px;
      color: ${props => props.theme.colors.neutral30};
      cursor: pointer;
    }
  }
  .secret-data-hint {
    background-color: ${props => props.theme.colors.primaryLight05};
    padding: ${props => props.theme.spacing.xs};
    font-weight: 600;
    width: fit-content;
    border-radius: 4px;
    margin-top: 0px;
  }
  .add-secret-data {
    margin-bottom: ${props => props.theme.spacing.medium};
  }
  .gitops-wrapper {
    padding-bottom: 0px;
  }
  .create-cta {
    display: flex;
    justify-content: end;
  }
`;
export const PreviewPRSection = styled.div`
  display: flex;
  justify-content: flex-end;
  padding: ${props => props.theme.spacing.small};
`;
