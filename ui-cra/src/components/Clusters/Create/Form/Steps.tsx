import React, {
  Dispatch,
  ReactElement,
  useCallback,
  useEffect,
  useState,
} from 'react';
import theme from 'weaveworks-ui-components/lib/theme';
import { ThemeProvider, createTheme } from '@material-ui/core/styles';
import { muiTheme } from '../../../../muiTheme';
import { FormStep } from './Step';

const localMuiTheme = createTheme({
  ...muiTheme,
  overrides: {
    ...muiTheme.overrides,
    MuiInputBase: {
      ...muiTheme.overrides?.MuiInputBase,
      root: {
        ...muiTheme.overrides?.MuiInputBase?.root,
        marginRight: `${theme.spacing.xs}`,
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
      asterisk: {
        display: 'none',
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

const FormSteps = {
  Box: (props: { properties: Property[] }) => {
    const [properties, setProperties] = useState<Property[]>([]);
    const [repeatChildrenVisible, setRepeatChildrenVisible] =
      useState<boolean>(false);
    const [childrenOccurences, setChildrenOccurences] =
      useState<
        {
          name: string;
          groupVisible: boolean;
          count: number;
        }[]
      >();

    const switchChildVisibility = useCallback(
      (childName: string) => {
        const updatedProperties = properties.map(property => {
          const updatedChildren = property.children.map(child => {
            if (child.props.name === childName && !child.props.firstOfAKind) {
              return React.cloneElement(child, {
                visible: !child.props.visible,
              });
            }
            return child;
          });
          property.children = updatedChildren;
          return property;
        });
        setProperties(updatedProperties);
      },
      [properties],
    );

    const getChildrenOccurences = useCallback(() => {
      const getChildrenNamesAndVisibility = properties.flatMap(property =>
        property.children.map(child => {
          return { name: child.props.name, visible: child.props.visible };
        }),
      );

      let childrenCountGroupVisibility: {
        name: string;
        groupVisible: boolean;
        count: number;
      }[] = [];

      getChildrenNamesAndVisibility.forEach(child => {
        const relevantChild = childrenCountGroupVisibility.find(
          c => c.name === child.name,
        );
        if (relevantChild) {
          relevantChild.count++;
          relevantChild.groupVisible = child.visible;
        } else {
          childrenCountGroupVisibility.push({
            name: child.name,
            count: 1,
            groupVisible: child.visible,
          });
        }
      });

      return childrenCountGroupVisibility;
    }, [properties]);

    useEffect(() => {
      setProperties(props.properties);
      setChildrenOccurences(getChildrenOccurences());
    }, [props.properties, getChildrenOccurences]);

    return (
      <ThemeProvider theme={localMuiTheme}>
        {properties.map((p, index) => {
          return (
            <FormStep
              key={index}
              step={p}
              active={p.active}
              clicked={p.clicked}
              setActiveStep={p.setActiveStep}
              childrenOccurences={childrenOccurences}
              switchChildVisibility={switchChildVisibility}
              repeatChildrenVisible={repeatChildrenVisible}
              setRepeatChildrenVisible={setRepeatChildrenVisible}
            />
          );
        })}
      </ThemeProvider>
    );
  },
};

export default FormSteps;
