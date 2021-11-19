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
import { Button } from 'weaveworks-ui-components';
import classNames from 'classnames';

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
  overflow: hidden;
  .form-group {
  }
  .step-child {
    display: flex;
    margin-bottom: ${theme.spacing.small};
    margin-right: ${theme.spacing.large};
    .step-child-btn {
      align-self: flex-end;
      height: 40px;
      overflow: hidden;
    }
  }
  @media (max-width: 768px) {
    flex-direction: column;
  }
`;

const SectionDivider = styled.div`
  margin-top: ${theme.spacing.medium};
`;

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
  className?: string;
  step?: Property;
  title?: string;
  active?: boolean;
  clicked?: boolean;
  setActiveStep?: Dispatch<React.SetStateAction<string | undefined>>;
  childrenOccurences?: { [key: string]: number }[];
  makeChildVisible?: (childName: string) => void;
}> = ({
  className,
  step,
  title,
  active,
  clicked,
  setActiveStep,
  childrenOccurences,
  makeChildVisible,
  children,
}) => {
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

  const handleClick = (childName: string) => {
    makeChildVisible && makeChildVisible(childName);
  };

  return (
    <Section ref={stepRef} className={className}>
      <Title name={title}>{step?.name || title}</Title>
      <Content>
        {step?.children.map((child, index) => {
          if (child.props.visible) {
            const occurences =
              (childrenOccurences && childrenOccurences[child.props.name]) || 0;

            console.log(index, child.props.name);

            return (
              <div key={index} className="step-child">
                {child}
                {occurences > 1 ? (
                  <Button
                    className={classNames(
                      'step-child-btn',
                      `step-child-btn-${child.props.name}`,
                    )}
                    title={child.props.name}
                    onClick={() => handleClick(child.props.name)}
                  >
                    Show &nbsp;
                    {occurences}
                    &nbsp; populated occurences
                  </Button>
                ) : null}
              </div>
            );
          }
          return null;
        })}
      </Content>
      {children}
      <SectionDivider>
        <Divider />
      </SectionDivider>
    </Section>
  );
};
