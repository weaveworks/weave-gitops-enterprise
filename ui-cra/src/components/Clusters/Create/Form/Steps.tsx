import React, {
  Dispatch,
  ReactElement,
  useCallback,
  useEffect,
  useState,
} from 'react';
import { FormStep } from './Step';
import { ChildrenOccurences } from '../../../../types/custom';

interface Property {
  name: string;
  active?: boolean;
  clicked?: boolean;
  setActiveStep?: Dispatch<React.SetStateAction<string | undefined>>;
  children: ReactElement[];
  addUserSelectedFields: (name: string) => void;
}

const FormSteps = {
  Box: (props: { properties: Property[] }) => {
    const [properties, setProperties] = useState<Property[]>([]);
    const [childrenOccurences, setChildrenOccurences] = useState<
      ChildrenOccurences[]
    >([]);

    const getChildrenOccurences = useCallback(() => {
      const getChildrenNamesAndVisibility = properties?.flatMap(property =>
        property.children.map(child => {
          return { name: child.props.name, visible: child.props.visible };
        }),
      );

      let childrenCountGroupVisibility: ChildrenOccurences[] = [];

      getChildrenNamesAndVisibility?.forEach(child => {
        const currentChild = childrenCountGroupVisibility.find(
          c => c.name === child.name,
        );
        if (currentChild) {
          currentChild.count++;
          currentChild.groupVisible = child.visible;
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
      <>
        {properties?.map((p, index) => {
          return (
            <FormStep
              key={index}
              step={p}
              active={p.active}
              clicked={p.clicked}
              setActiveStep={p.setActiveStep}
              childrenOccurences={childrenOccurences}
              addUserSelectedFields={p.addUserSelectedFields}
            />
          );
        })}
      </>
    );
  },
};

export default FormSteps;
