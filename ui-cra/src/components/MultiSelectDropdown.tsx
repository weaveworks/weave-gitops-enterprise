import React, { Dispatch, FC, useState } from 'react';
import Checkbox from '@material-ui/core/Checkbox';
import ListItemIcon from '@material-ui/core/ListItemIcon';
import ListItemText from '@material-ui/core/ListItemText';
import MenuItem from '@material-ui/core/MenuItem';
import FormControl from '@material-ui/core/FormControl';
import Select from '@material-ui/core/Select';
import { makeStyles } from '@material-ui/core/styles';
import { Profile } from '../types/custom';
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
  items: any[];
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
                color: GitOpsBlue,
              }}
            />
          </ListItemIcon>
          <ListItemText primary="Select All" />
        </MenuItem>
        {itemsNames.map(item => (
          <MenuItem key={item} value={item}>
            <ListItemIcon>
              <Checkbox
                checked={selected.indexOf(item) > -1}
                style={{
                  color: GitOpsBlue,
                }}
              />
            </ListItemIcon>
            <ListItemText primary={item} />
          </MenuItem>
        ))}
      </Select>
    </FormControl>
  );
};

export default MultiSelectDropdown;
