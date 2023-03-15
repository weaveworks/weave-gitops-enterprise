import { find, flattenDepth, isArray, map, noop } from 'lodash';
import PropTypes from 'prop-types';
import React from 'react';
import styled from 'styled-components';

import { spacing } from '../../theme/selectors';

const WIDTH = '256px';
const HEIGHT_NUMBER = 36;
const HEIGHT = `${HEIGHT_NUMBER}px`;

const Item = styled.div`
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  padding-right: ${spacing('medium')};
  padding-left: ${spacing('small')};
  cursor: pointer;
  ${props =>
    props.disabled &&
    `
    cursor: not-allowed;
    color: ${props.theme.colors.gray600};
    background-color: ${props.theme.colors.gray50};
  `};
`;

const Popover = styled.div`
  position: absolute;
  background-color: ${props => props.theme.colors.white};
  border: 1px solid ${props => props.theme.colors.gray200};
  border-radius: ${props => props.theme.borderRadius.soft};
  z-index: ${props => props.theme.layers.dropdown};
  box-shadow: ${props => props.theme.boxShadow.light};
  margin-top: 4px;
  width: ${props => props.width}px;
  /* +10 to account for list top padding.
   * Multiply by an 0.5 to visually slice the last item in half to more clearly show that the
   * dropdown can be scrolled. */
  max-height: ${HEIGHT_NUMBER * 10.5 + 10}px;
  overflow: auto;
  box-sizing: border-box;
  /* padding: ${spacing('xxs')} 0; */
`;

const Overlay = styled.div`
  z-index: 4;
  position: fixed;
  top: 0;
  bottom: 0;
  right: 0;
  left: 0;
`;

const ItemWrapper = styled(Item)`
  line-height: ${HEIGHT};
  ${props => props.selected && `color: blue`}
  min-height: ${HEIGHT};
  &:hover:not([disabled]) {
    background-color: gray;
  }
`;

const Divider = styled.div`
  margin: 6px 0;
  border-bottom: 1px solid gray;
`;

const SelectedItem = styled(Item)`
  height: ${HEIGHT};
  box-sizing: border-box;
  border-radius: ${props => props.theme.borderRadius.soft};
  background-color: ${props => props.theme.colors.white};
  border: 1px solid gray;
  display: flex;
  ${Item} {
    padding: 0;
  }
  div:last-child {
    margin-left: auto;
  }
  ${props => props.disabled && `background-color: gray;`};
`;

const SelectedItemIcon = styled.span`
  position: absolute;
  right: 10px;
  top: 50%;
  transform: translateY(-50%);
`;

const StyledDropdown = component => styled(component)`
  height: ${HEIGHT};
  line-height: 34px;
  position: relative;
  width: ${props => props.width || WIDTH};
  box-sizing: border-box;
`;

const DefaultToggleView = ({ onClick, disabled, selectedLabel }) => (
  <SelectedItem
    className="dropdown-toggle"
    onClick={onClick}
    disabled={disabled}
  >
    <Item disabled={disabled}>{selectedLabel}</Item>
    <SelectedItemIcon className="fa fa-caret-down" />
  </SelectedItem>
);

/**
 * A selectable drop-down menu.
 * ```javascript
 *const items = [
 *  {
 *    value: 'first-thing',
 *    label: 'First Thing',
 *  },
 *  {
 *    value: 'second-thing',
 *    label: 'Second Thing',
 *  },
 * ];
 *
 * <Dropdown items={items} />
 * ```
 *
 * You  may also add `null` for dividers or provide groups that will be separated
 * ```javascript
 *const items = [
 *  [
 *    {
 *      value: 'first-thing',
 *      label: 'First Thing',
 *    },
 *    {
 *      value: 'second-thing',
 *      label: 'Second Thing',
 *    },
 *  ],
 *  [
 *    {
 *      value: 'another-thing',
 *      label: 'Another Thing',
 *    },
 *  ]
 * ];
 * ```
 */
class Dropdown extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.state = {
      isOpen: false,
    };

    this.element = React.createRef();

    this.handleChange = this.handleChange.bind(this);
    this.handleClick = this.handleClick.bind(this);
    this.handleBgClick = this.handleBgClick.bind(this);
  }

  handleChange(ev, value, label) {
    this.setState({ isOpen: false });
    if (this.props.onChange) {
      this.props.onChange(ev, value, label);
    }
  }

  handleClick() {
    if (!this.props.disabled) {
      this.setState({ isOpen: true });
    }
  }

  handleBgClick() {
    this.setState({ isOpen: false });
  }

  divide(items) {
    if (!items || !isArray(items[0])) {
      return items;
    }

    return flattenDepth(
      items.map(it => [null, it]),
      2,
    ).slice(1);
  }

  render() {
    const { items, value, className, placeholder, disabled } = this.props;
    const { isOpen } = this.state;
    const divided = this.divide(items);
    // If nothing is selected, use the placeholder, else use the first item.
    const currentItem =
      find(divided, i => i && i.value === value) ||
      (placeholder
        ? { label: placeholder, value: null }
        : divided && divided[0]);
    const label =
      currentItem && (currentItem.selectedLabel || currentItem.label);
    const Component = this.props.withComponent;

    return (
      <div className={className} title={label} ref={this.element}>
        <Component
          selectedLabel={label}
          disabled={disabled}
          onClick={this.handleClick}
        />
        {isOpen && (
          <div>
            <Overlay onClick={this.handleBgClick} />
            <Popover
              className="dropdown-popover"
              width={this.element.current.offsetWidth}
            >
              {map(divided, (item, index) =>
                item ? (
                  <ItemWrapper
                    className="dropdown-item"
                    key={item.value}
                    disabled={item.disabled}
                    onClick={
                      item.disabled
                        ? undefined
                        : ev => this.handleChange(ev, item.value, item.label)
                    }
                    selected={item.value === value}
                    title={
                      item && typeof item.label === 'string' ? item.label : ''
                    }
                  >
                    {item.label}
                  </ItemWrapper>
                ) : (
                  <Divider key={index} />
                ),
              )}
            </Popover>
          </div>
        )}
      </div>
    );
  }
}

const itemPropType = PropTypes.shape({
  disabled: PropTypes.bool,
  label: PropTypes.node,
  selectedLabel: PropTypes.node,
  value: PropTypes.oneOfType([PropTypes.string, PropTypes.number]),
});

Dropdown.propTypes = {
  /**
   * Disables the component if true
   */
  disabled: PropTypes.bool,
  /**
   * Array of items (or groups of items) that will be selectable.
   * - `value` should be an internal value,
   * - `label` is what will be displayed to the user.
   * - `selectedLabel` (optional) what will be displayed in the collapsed state of the dropdown. If
   *   omitted `label` will be used.
   */
  items: PropTypes.arrayOf(
    PropTypes.oneOfType([PropTypes.arrayOf(itemPropType), itemPropType]),
  ).isRequired,
  /**
   * A handler function that will run when a value is selected.
   */
  onChange: PropTypes.func,
  /**
   * The initial text that will be display before a user selects an item.
   */
  placeholder: PropTypes.string,
  /**
   * The value of the currently selected item. This much match a value in the `items` prop.
   * If no value is provided, the first elements's value will be used.
   */
  value: PropTypes.oneOfType([PropTypes.string, PropTypes.number]),

  /**
   * Pass a custom width css value
   */
  width: PropTypes.string,

  /**
   * A custom component to replace the default toggle view.
   * The properties `selectedLabel` and `onClick` are provided. `onClick` needs to be incorporated
   * to make the dropdown list toggle.
   */
  withComponent: PropTypes.func,
};

Dropdown.defaultProps = {
  disabled: false,
  onChange: noop,
  placeholder: '',
  value: '',
  width: WIDTH,
  withComponent: DefaultToggleView,
};

export default StyledDropdown(Dropdown);
