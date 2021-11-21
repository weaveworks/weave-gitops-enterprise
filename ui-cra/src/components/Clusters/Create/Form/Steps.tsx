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
    const [properties, setProperties] = useState<Property[]>(props.properties);

    const makeChildVisible = useCallback(
      (childName: string) => {
        const updatedProperties = properties.map(property => {
          const updatedChildren = property.children.map(child => {
            if (child.props.name === childName) {
              return React.cloneElement(child, {
                visible: true,
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
      const getChildrenNames = properties.flatMap(property =>
        property.children.map(child => child.props.name),
      );

      return getChildrenNames.reduce((namesWithCount, name) => {
        Object.keys(namesWithCount).includes(name)
          ? namesWithCount[name]++
          : (namesWithCount[name] = 1);
        return namesWithCount;
      }, {});
    }, [properties]);

    useEffect(() => {
      console.log('updating props by typing in form triggers rerender');
      setProperties(props.properties);
    }, [props.properties]);

    console.log('i am rerendering');

    const childrenOccurences = getChildrenOccurences();

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
              makeChildVisible={makeChildVisible}
            />
          );
        })}
      </ThemeProvider>
    );
  },
};

export default FormSteps;
