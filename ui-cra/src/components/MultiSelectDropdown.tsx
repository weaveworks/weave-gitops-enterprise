import React, { Dispatch, FC, useState } from 'react';
import Checkbox from '@material-ui/core/Checkbox';
import ListItemIcon from '@material-ui/core/ListItemIcon';
import ListItemText from '@material-ui/core/ListItemText';
import MenuItem from '@material-ui/core/MenuItem';
import FormControl from '@material-ui/core/FormControl';
import Select from '@material-ui/core/Select';
import { makeStyles } from '@material-ui/core/styles';
import { Profile } from '../types/custom';
<<<<<<< HEAD
import { GitOpsBlue } from './../muiTheme';
=======
>>>>>>> a45ff383b9ccee79e24918725ec38a422e4e6363

const useStyles = makeStyles(theme => ({
  formControl: {
    margin: theme.spacing(1),
    width: 300,
  },
  indeterminateColor: {
<<<<<<< HEAD
    color: GitOpsBlue,
  },
  downloadBtn: {
    color: GitOpsBlue,
=======
    color: '#00B3EC',
  },
  downloadBtn: {
    color: '#00B3EC',
>>>>>>> a45ff383b9ccee79e24918725ec38a422e4e6363
    padding: '0px',
  },
}));

const MultiSelectDropdown: FC<{
  items: any[];
<<<<<<< HEAD
  onSelectItems: Dispatch<React.SetStateAction<Profile[]>>;
}> = ({ items, onSelectItems }) => {
  const classes = useStyles();
  const [selected, setSelected] = useState<any[]>([]);
  const isAllSelected = items.length > 0 && selected.length === items.length;

  const getItemsFromNames = (names: string[]) =>
    items.filter(item => names.find(name => item.name === name));

  const handleChange = (event: React.ChangeEvent<any>) => {
    const value = event.target.value;
    if (value[value.length - 1] === 'all') {
      const selectedItems = selected.length === items.length ? [] : items;
      setSelected(selectedItems);
      onSelectItems(selectedItems);
      return;
    }
    setSelected(value);
    onSelectItems(getItemsFromNames(value));
=======
  onSelectProfiles?: Dispatch<React.SetStateAction<Profile[]>>;
}> = ({ items, onSelectProfiles }) => {
  const classes = useStyles();
  const [selected, setSelected] = useState<string[]>([]);
  const isAllSelected = items.length > 0 && selected.length === items.length;

  const itemsNames = items.map(item => item.name);

  const getItemsFromNames = (names: string[]) =>
    items.filter(item => names.find(name => item.name === name));

  const handleChange = (event: any) => {
    const value = event.target.value;
    if (value[value.length - 1] === 'all') {
      const selectedItems = selected.length === itemsNames.length ? [] : items;
      const selectedItemsNames =
        selected.length === itemsNames.length ? [] : itemsNames;
      setSelected(selectedItemsNames);
      onSelectProfiles && onSelectProfiles(selectedItems);
      return;
    }
    setSelected(value);
    onSelectProfiles && onSelectProfiles(getItemsFromNames(value));
>>>>>>> a45ff383b9ccee79e24918725ec38a422e4e6363
  };

  return (
    <FormControl className={classes.formControl}>
      <Select
        labelId="mutiple-select-label"
        multiple
        value={selected}
        onChange={handleChange}
        renderValue={(selected: any) => selected.join(', ')}
      >
        <MenuItem value="all">
          <ListItemIcon>
            <Checkbox
              classes={{ indeterminate: classes.indeterminateColor }}
              checked={isAllSelected}
              indeterminate={
                selected.length > 0 && selected.length < items.length
              }
              style={{
<<<<<<< HEAD
                color: GitOpsBlue,
=======
                color: '#00B3EC',
>>>>>>> a45ff383b9ccee79e24918725ec38a422e4e6363
              }}
            />
          </ListItemIcon>
          <ListItemText primary="Select All" />
        </MenuItem>
<<<<<<< HEAD
        {items.map(item => {
          const itemName = item.name;
          return (
            <MenuItem key={itemName} value={itemName}>
              <ListItemIcon>
                <Checkbox
                  checked={selected.indexOf(itemName) > -1}
                  style={{
                    color: GitOpsBlue,
                  }}
                />
              </ListItemIcon>
              <ListItemText primary={itemName} />
            </MenuItem>
          );
        })}
=======
        {itemsNames.map(item => (
          <MenuItem key={item} value={item}>
            <ListItemIcon>
              <Checkbox
                checked={selected.indexOf(item) > -1}
                style={{
                  color: '#00B3EC',
                }}
              />
            </ListItemIcon>
            <ListItemText primary={item} />
          </MenuItem>
        ))}
>>>>>>> a45ff383b9ccee79e24918725ec38a422e4e6363
      </Select>
    </FormControl>
  );
};

export default MultiSelectDropdown;
