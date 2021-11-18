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

const FormSteps = {
  Box: (props: { properties: Property[] }) => {
    const [properties, setProperties] = useState<Property[]>([]);

    const makeChildVisible = (childName: string) => {
      console.log(childName);
    };

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
      setProperties(props.properties);
    }, [props.properties]);

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
