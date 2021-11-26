import React, {
  Dispatch,
  FC,
  ReactElement,
  Ref,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from 'react';
import Divider from '@material-ui/core/Divider';
import styled from 'styled-components';
import theme from 'weaveworks-ui-components/lib/theme';
import { Button } from 'weaveworks-ui-components';
import { GitOpsBlue } from '../../../../muiTheme';
import classNames from 'classnames';
import { ChildrenOccurences } from '../../../../types/custom';

const Section = styled.div`
  padding-bottom: ${theme.spacing.medium};
`;

const Title = styled.div<{ name?: string }>`
  display: flex;
  align-items: center;
  padding-bottom: ${theme.spacing.small};
  font-size: ${theme.fontSizes.large};
  span {
    margin-left: ${theme.spacing.small};
    font-size: ${theme.fontSizes.tiny};
    color: rgba(0, 0, 0, 0.54);
    font-weight: 400;
  }
`;

const Content = styled.div`
  display: flex;
  overflow: hidden;
  .step-child {
    display: flex;
    margin-bottom: ${theme.spacing.small};
    margin-right: ${theme.spacing.large};
    .step-child-btn {
      align-self: flex-end;
      height: 40px;
      overflow: hidden;
    }
    span {
      color: ${GitOpsBlue};
      font-weight: 600;
    }
  }
  .step-child-disabled {
    input {
      background-color: #f5f5f5;
      cursor: not-allowed;
    }
    input:focus-within {
      pointer-events: none;
    }
    div[role='button'] {
      background-color: #f5f5f5;
      pointer-events: none;
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
  childrenOccurences?: ChildrenOccurences[];
  addUserSelectedFields?: (name: string) => void;
}> = ({
  className,
  step,
  title,
  active,
  clicked,
  setActiveStep,
  childrenOccurences,
  children,
  addUserSelectedFields,
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

  const handleClick = useCallback(
    (childName: string) => {
      addUserSelectedFields && addUserSelectedFields(childName);
    },
    [addUserSelectedFields],
  );

  const hiddenFieldsPerStep =
    step?.children.filter(child => !child.props.visible).length || 0;

  return (
    <Section ref={stepRef} className={className}>
      <Title name={title}>
        {step?.name || title}&nbsp;
        {hiddenFieldsPerStep !== 0 ? (
          <span>
            (&nbsp;{hiddenFieldsPerStep}&nbsp;
            {hiddenFieldsPerStep > 1 ? 'HIDDEN FIELDS' : 'HIDDEN FIELD'} )
          </span>
        ) : null}
      </Title>
      <Content>
        {step?.children.map((child, index) => {
          if (child.props.visible) {
            const childOccurences = childrenOccurences?.find(
              c => c.name === child.props.name,
            ) as ChildrenOccurences;
            return (
              <div
                key={index}
                className={classNames(
                  'step-child',
                  child.props.firstOfAKind === false
                    ? 'step-child-disabled'
                    : '',
                )}
              >
                {child}
                {childOccurences?.count > 1 && child.props.firstOfAKind ? (
                  <Button
                    type="button"
                    className="step-child-btn"
                    onClick={() => handleClick(child.props.name)}
                  >
                    {childOccurences?.groupVisible ? 'Hide' : 'Show'}&nbsp;
                    <span>{childOccurences.count}</span> populated fields
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
