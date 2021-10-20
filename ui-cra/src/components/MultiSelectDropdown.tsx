import React, { FC, useState } from 'react';
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

const MultiSelectDropdown: FC<{ items: any[]; onSelectProfiles: any }> = ({
  items,
  onSelectProfiles,
}) => {
  const classes = useStyles();
  const [selected, setSelected] = useState<string[]>([]);
  const isAllSelected = items.length > 0 && selected.length === items.length;

  const itemsNames = items.map(item => item.name);

  const handleChange = (event: any) => {
    const value = event.target.value;
    if (value[value.length - 1] === 'all') {
      setSelected(selected.length === itemsNames.length ? [] : itemsNames);
      onSelectProfiles(selected.length === itemsNames.length ? [] : itemsNames);
      return;
    }
    setSelected(value);
    onSelectProfiles(value);
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
                color: '#00B3EC',
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
                  color: '#00B3EC',
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
