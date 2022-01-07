import React, { FC, useEffect, useState } from 'react';
import Checkbox from '@material-ui/core/Checkbox';
import ListItemIcon from '@material-ui/core/ListItemIcon';
import ListItemText from '@material-ui/core/ListItemText';
import MenuItem from '@material-ui/core/MenuItem';
import FormControl from '@material-ui/core/FormControl';
import Select from '@material-ui/core/Select';
import { makeStyles } from '@material-ui/core/styles';
import { theme as weaveTheme } from '@weaveworks/weave-gitops';

const useStyles = makeStyles(theme => ({
  formControl: {
    margin: theme.spacing(1),
    width: 300,
  },
  indeterminateColor: {
    color: weaveTheme.colors.primary,
  },
  downloadBtn: {
    color: weaveTheme.colors.primary,
    padding: '0px',
  },
}));

const MultiSelectDropdown: FC<{
  items: any[];
  onSelectItems: (items: any[]) => void;
}> = ({ items, onSelectItems }) => {
  const classes = useStyles();
  const [selected, setSelected] = useState<any[]>([]);
  const onlyRequiredProfiles =
    items.filter(item => item.required === true).length === items.length;
  const isAllSelected =
    items.length > 0 &&
    (selected.length === items.length || onlyRequiredProfiles);

  const getItemsFromNames = (names: string[]) =>
    items.filter(item => names.find(name => item.name === name));

  const getNamesFromItems = (items: any[]) => items.map(item => item.name);

  const handleChange = (event: React.ChangeEvent<any>) => {
    const value = event.target.value;
    if (value[value.length - 1] === 'all') {
      const selectedItems = selected.length === items.length ? [] : items;
      setSelected(getNamesFromItems(selectedItems));
      onSelectItems(selectedItems);
      return;
    }
    setSelected(value);
    onSelectItems(getItemsFromNames(value));
  };

  useEffect(
    () =>
      setSelected(
        getNamesFromItems(items.filter(item => item.required === true)),
      ),

    [items],
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
        <MenuItem value="all" disabled={onlyRequiredProfiles}>
          <ListItemIcon>
            <Checkbox
              classes={{ indeterminate: classes.indeterminateColor }}
              checked={isAllSelected}
              indeterminate={
                selected.length > 0 && selected.length < items.length
              }
              style={{
                color: weaveTheme.colors.primary,
              }}
            />
          </ListItemIcon>
          <ListItemText primary="Select All" />
        </MenuItem>
        {items.map(item => {
          const itemName = item.name;
          return (
            <MenuItem key={itemName} value={itemName} disabled={item.required}>
              <ListItemIcon>
                <Checkbox
                  checked={
                    item.required === true || selected.indexOf(itemName) > -1
                  }
                  style={{
                    color: weaveTheme.colors.primary,
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
