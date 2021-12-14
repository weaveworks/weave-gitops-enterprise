import React, { FC, useEffect, useState } from 'react';
import Checkbox from '@material-ui/core/Checkbox';
import ListItemIcon from '@material-ui/core/ListItemIcon';
import ListItemText from '@material-ui/core/ListItemText';
import MenuItem from '@material-ui/core/MenuItem';
import FormControl from '@material-ui/core/FormControl';
import Select from '@material-ui/core/Select';
import { makeStyles } from '@material-ui/core/styles';
import { GitOpsBlue } from './../muiTheme';

const useStyles = makeStyles(theme => ({
  formControl: {
    margin: theme.spacing(1),
    width: 300,
  },
  indeterminateColor: {
    color: GitOpsBlue,
  },
  downloadBtn: {
    color: GitOpsBlue,
    padding: '0px',
  },
}));

const MultiSelectDropdown: FC<{
  allItems: any[];
  preSelectedItems: any[];
  onSelectItems: (items: any[]) => void;
}> = ({ allItems, preSelectedItems, onSelectItems }) => {
  const classes = useStyles();
  const [selected, setSelected] = useState<any[]>([]);
  const onlyRequiredItems =
    allItems.filter(item => item.required === true).length === allItems.length;
  const isAllSelected =
    allItems.length > 0 &&
    (selected.length === allItems.length || onlyRequiredItems);

  const getItemsFromNames = (names: string[]) =>
    allItems.filter(item => names.find(name => item.name === name));

  const getNamesFromItems = (items: any[]) => items.map(item => item.name);

  const handleChange = (event: React.ChangeEvent<any>) => {
    const value = event.target.value;
    if (value[value.length - 1] === 'all') {
      const selectedItems = selected.length === allItems.length ? [] : allItems;
      setSelected(getNamesFromItems(selectedItems));
      onSelectItems(selectedItems);
      return;
    }
    setSelected(value);
    onSelectItems(getItemsFromNames(value));
  };

  useEffect(
    () => setSelected(getNamesFromItems(preSelectedItems)),
    [preSelectedItems],
  );

  return (
    <FormControl className={classes.formControl}>
      <Select
        labelId="mutiple-select-label"
        multiple
        value={selected}
        onChange={handleChange}
        renderValue={(selected: any) => selected.join(', ')}
      >
        <MenuItem value="all" disabled={onlyRequiredItems}>
          <ListItemIcon>
            <Checkbox
              classes={{ indeterminate: classes.indeterminateColor }}
              checked={isAllSelected}
              indeterminate={
                selected.length > 0 && selected.length < allItems.length
              }
              style={{
                color: GitOpsBlue,
              }}
            />
          </ListItemIcon>
          <ListItemText primary="Select All" />
        </MenuItem>
        {allItems.map(item => {
          const itemName = item.name;
          return (
            <MenuItem key={itemName} value={itemName} disabled={item.required}>
              <ListItemIcon>
                <Checkbox
                  checked={
                    item.required === true || selected.indexOf(itemName) > -1
                  }
                  style={{
                    color: GitOpsBlue,
                  }}
                />
              </ListItemIcon>
              <ListItemText primary={itemName} />
            </MenuItem>
          );
        })}
      </Select>
    </FormControl>
  );
};

export default MultiSelectDropdown;
