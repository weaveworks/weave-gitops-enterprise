import React, { FC, ReactElement, Ref, useEffect, useRef } from 'react';
import Divider from '@material-ui/core/Divider';
import styled from 'styled-components';
import theme from 'weaveworks-ui-components/lib/theme';
import { ThemeProvider, createMuiTheme } from '@material-ui/core/styles';
import { muiTheme } from '../../../../muiTheme';

const Section = styled.div`
  padding-bottom: ${theme.spacing.medium};
`;

const Title = styled.div`
  padding-bottom: ${theme.spacing.small};
  font-size: ${theme.fontSizes.large};
`;

const Content = styled.div`
  display: flex;
  @media (max-width: 768px) {
    flex-direction: column;
  }
`;

const SectionDivider = styled.div`
  margin-top: ${theme.spacing.medium};
`;

const localMuiTheme = createMuiTheme({
  ...muiTheme,
  overrides: {
    ...muiTheme.overrides,
    MuiInputBase: {
      ...muiTheme.overrides?.MuiInputBase,
      root: {
        ...muiTheme.overrides?.MuiInputBase?.root,
        marginRight: `${theme.spacing.xxl}`,
        marginBottom: `${theme.spacing.xs}`,
      },
      input: {
        ...muiTheme.overrides?.MuiInputBase?.input,
        '&:focus': {
          border: 'none',
        },
      },
    },
    MuiInputLabel: {
      formControl: {
        ...muiTheme.overrides?.MuiInputLabel?.formControl,
        fontSize: `${theme.fontSizes.tiny}`,
      },
    },
  },
});

interface Property {
  name: string;
  active?: boolean;
  children: ReactElement[];
}

const FormStep: FC<{
  step: Property;
  active?: boolean;
}> = ({ step, active }) => {
  const stepRef: Ref<HTMLDivElement> = useRef(null);

  useEffect(() => {
    if (active && stepRef.current) {
      stepRef.current.scrollIntoView({ behavior: 'smooth', block: 'center' });
    }
  }, [active]);

  return (
    <Section ref={stepRef}>
      <Title>{step.name}</Title>
      <Content>{step.children}</Content>
      <SectionDivider>
        <Divider />
      </SectionDivider>
    </Section>
  );
};

const FormSteps = {
  box: (props: { properties: Property[] }) => (
    <ThemeProvider theme={localMuiTheme}>
      {props.properties.map((p, index) => {
        return <FormStep key={index} step={p} active={p.active} />;
      })}
    </ThemeProvider>
  ),
};

export default FormSteps;
