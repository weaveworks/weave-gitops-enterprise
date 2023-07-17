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
`;
export const PreviewPRSection = styled.div`
  display: flex;
  justify-content: flex-end;
  padding: ${props => props.theme.spacing.small};
`;
