import Checkbox from '@material-ui/core/Checkbox';
import FormControl from '@material-ui/core/FormControl';
import ListItemIcon from '@material-ui/core/ListItemIcon';
import ListItemText from '@material-ui/core/ListItemText';
import MenuItem from '@material-ui/core/MenuItem';
import Select from '@material-ui/core/Select';
import { makeStyles } from '@material-ui/core/styles';
import React, { FC, useEffect, useState } from 'react';
import styled from 'styled-components';

const useStyles = makeStyles(theme => ({
  formControl: {
    margin: theme.spacing(1),
    width: 300,
  },
  downloadBtn: {
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
        MenuProps={{
          anchorOrigin: {
            vertical: 'bottom',
            horizontal: 'left',
          },
          transformOrigin: {
            vertical: 'top',
            horizontal: 'left',
          },
          getContentAnchorEl: null,
        }}
      >
        <MenuItem value="all" disabled={onlyRequiredItems}>
          <ListItemIcon>
            <Checkbox
              checked={isAllSelected}
              indeterminate={
                selected.length > 0 && selected.length < allItems.length
              }
              color="primary"
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
                  color="primary"
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

export default styled(MultiSelectDropdown)``;
