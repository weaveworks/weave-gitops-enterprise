import React, { useState } from 'react';
import Checkbox from '@material-ui/core/Checkbox';
import ListItemIcon from '@material-ui/core/ListItemIcon';
import ListItemText from '@material-ui/core/ListItemText';
import MenuItem from '@material-ui/core/MenuItem';
import FormControl from '@material-ui/core/FormControl';
import Select from '@material-ui/core/Select';
import { makeStyles } from '@material-ui/core/styles';

const useStyles = makeStyles(theme => ({
  formControl: {
    margin: theme.spacing(1),
    width: 300,
  },
  indeterminateColor: {
    color: '#00B3EC',
  },
}));

const options = ['Profile 1', 'Profile 2', 'Profile 3'];

const MultiSelectDropdown = () => {
  const classes = useStyles();
  const [selected, setSelected] = useState<string[]>([]);
  const isAllSelected =
    options.length > 0 && selected.length === options.length;

  const handleChange = (event: any) => {
    const value = event.target.value;
    if (value[value.length - 1] === 'all') {
      setSelected(selected.length === options.length ? [] : options);
      return;
    }
    setSelected(value);
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
                selected.length > 0 && selected.length < options.length
              }
              style={{
                color: '#00B3EC',
              }}
            />
          </ListItemIcon>
          <ListItemText primary="Select All" />
        </MenuItem>
        {options.map(option => (
          <MenuItem key={option} value={option}>
            <ListItemIcon>
              <Checkbox
                checked={selected.indexOf(option) > -1}
                style={{
                  color: '#00B3EC',
                }}
              />
            </ListItemIcon>
            <ListItemText primary={option} />
          </MenuItem>
        ))}
      </Select>
    </FormControl>
  );
};

export default MultiSelectDropdown;
