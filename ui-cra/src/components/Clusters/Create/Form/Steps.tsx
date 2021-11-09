import React, {
  Dispatch,
  FC,
  ReactElement,
  Ref,
  useEffect,
  useMemo,
  useRef,
  useState,
} from 'react';
import Divider from '@material-ui/core/Divider';
import styled from 'styled-components';
import theme from 'weaveworks-ui-components/lib/theme';
import { ThemeProvider, createTheme } from '@material-ui/core/styles';
import { muiTheme } from '../../../../muiTheme';

const Section = styled.div`
  padding-bottom: ${theme.spacing.medium};
`;

const Title = styled.div<{ name?: string }>`
  padding-bottom: ${theme.spacing.small};
  font-size: ${theme.fontSizes.large};
  font-family: ${theme.fontFamilies.monospace};
`;

const Content = styled.div`
  display: flex;
  flex-wrap: wrap;
  overflow: hidden;
  @media (max-width: 768px) {
    flex-direction: column;
  }
`;

const SectionDivider = styled.div`
  margin-top: ${theme.spacing.medium};
`;

const localMuiTheme = createTheme({
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
        minWidth: '155px',
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
      shrink: {
        transform: 'none',
      },
    },
    MuiSelect: {
      select: {
        ...muiTheme.overrides?.MuiSelect?.select,
        minWidth: '155px',
      },
    },
  },
});

interface Property {
  name: string;
  active?: boolean;
  clicked?: boolean;
  setActiveStep?: Dispatch<React.SetStateAction<string | undefined>>;
  children: ReactElement[];
}

export const useOnScreen = (ref: { current: HTMLDivElement | null }) => {
  const [isIntersecting, setIntersecting] = useState(false);

  const observer = useMemo(
    () =>
      new IntersectionObserver(
        ([entry]) => setIntersecting(entry.isIntersecting),
        { rootMargin: '-50% 0px' },
      ),
    [],
  );

  useEffect(() => {
    if (ref.current) {
      observer.observe(ref.current);
      return () => {
        observer.disconnect();
      };
    }
  }, [observer, ref]);

  return isIntersecting;
};

export const FormStep: FC<{
  step?: Property;
  title?: string;
  active?: boolean;
  clicked?: boolean;
  setActiveStep?: Dispatch<React.SetStateAction<string | undefined>>;
}> = ({ step, title, active, clicked, setActiveStep, children }) => {
  const stepRef: Ref<HTMLDivElement> = useRef<HTMLDivElement>(null);

  const onScreen = useOnScreen(stepRef);

  useEffect(() => {
    if (clicked) {
      stepRef?.current?.scrollIntoView({
        behavior: 'smooth',
        block: 'center',
      });
    }
  }, [clicked]);

  useEffect(() => {
    setTimeout(() => {
      if (!active && onScreen) {
        setActiveStep && setActiveStep(step?.name || title);
      }
    }, 500);
  }, [active, setActiveStep, onScreen, step?.name, title]);

  console.log(step?.children);

  return (
    <Section ref={stepRef}>
      <Title name={title}>{step?.name || title}</Title>
      {step?.children ? <Content>{step?.children}</Content> : null}
      {children}
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
        return (
          <FormStep
            key={index}
            step={p}
            active={p.active}
            clicked={p.clicked}
            setActiveStep={p.setActiveStep}
          />
        );
      })}
    </ThemeProvider>
  ),
};

export default FormSteps;
