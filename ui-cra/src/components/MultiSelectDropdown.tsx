import React, { FC, useState } from 'react';
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
  items: any[];
  onSelectItems: (items: any[]) => void;
}> = ({ items, onSelectItems }) => {
  const classes = useStyles();
  const [selected, setSelected] = useState<any[]>([]);
  const isAllSelected =
    items.length > 0 &&
    (selected.length === items.length ||
      items.filter(item => item.required === true).length === items.length);

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
        {items.map(item => {
          const itemName = item.name;
          return (
            <MenuItem key={itemName} value={itemName}>
              <ListItemIcon>
                <Checkbox
                  checked={
                    item.required === true || selected.indexOf(itemName) > -1
                  }
                  style={{
                    color: GitOpsBlue,
                  }}
                  disabled={item.required}
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
