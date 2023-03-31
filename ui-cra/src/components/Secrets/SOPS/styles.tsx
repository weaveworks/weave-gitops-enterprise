import { theme } from '@weaveworks/weave-gitops';
import styled from 'styled-components';

const { xs, small, medium } = theme.spacing;

const { neutral30, neutral20, neutral10, primary10, primaryLight05 } =
  theme.colors;

export const FormWrapper = styled.form`
  .group-section {
    width: 100%;
    border-bottom: 1px dashed ${neutral20};
    .form-group {
      display: flex;
      flex-direction: column;
    }
    .form-section {
      width: 40%;
      .Mui-disabled {
        background: ${neutral10} !important;
        border-color: ${neutral20} !important;
      }
    }
    .MuiRadio-colorSecondary.Mui-checked {
      color: ${primary10};
    }
    h2 {
      font-size: 20px;
      margin-bottom: ${xs};
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
    .remove-icon {
      margin-top: 25px;
      color: ${neutral30};
      cursor: pointer;
    }
  }
  .secret-data-hint {
    background-color: ${primaryLight05};
    padding: ${xs};
    font-weight: 600;
    width: fit-content;
    border-radius: 4px;
    margin-top: 0px;
  }
  .add-secret-data {
    margin-bottom: ${medium};
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
  padding: ${small};
`;
