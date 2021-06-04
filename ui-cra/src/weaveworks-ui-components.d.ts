// Type definitions for weaveworks-ui-components v0.21.2
// Project: ui-components
// Definitions by: Dimitri Mitropoulos github.com/dimitropoulos

/*
see: https://github.com/weaveworks/ui-components/tree/master/src/theme

This file and its typings only exist as of 2019-08-22 a temporary compatability layer since weaveworks/ui-components does not yet have types and types are required by this codebase.

Any changes to the schema of the them in ui-componenets will have to be synced here until ui-components can provide a proper type file of its own.
*/

/* UTILS */
type ValueOf<T> = T[keyof T];
/* UTILS */

interface CustomProperties {
  hoverBackground: HexColor;
  hoverColor: HexColor;
}

interface Atoms {
  Alert: {
    error: {
      background: Colors['orange600'];
      color: Colors['white'];
    };
    info: {
      background: Colors['blue600'];
      color: Colors['white'];
    };
    success: {
      background: Colors['green500'];
      color: Colors['white'];
    };
    warning: {
      background: Colors['yellow500'];
      color: Colors['white'];
    };
  };
  Button: {
    danger: {
      background: Colors['orange600'];
      color: Colors['white'];
      hoverBackground: Colors['orange700'];
      hoverColor: Colors['white'];
    };
    default: {
      background: Colors['white'];
      color: Colors['black'];
      hoverBackground: Colors['gray50'];
      hoverColor: Colors['purple800'];
    };
    disabled: {
      background: Colors['gray50'];
      color: Colors['gray600'];
      hoverBackground: Colors['gray50'];
      hoverColor: Colors['gray600'];
    };
    primary: {
      background: Colors['blue700'];
      color: Colors['white'];
      hoverBackground: Colors['blue800'];
      hoverColor: Colors['white'];
    };
  };
}

interface BorderRadius {
  circle: '50%';
  none: '0';
  soft: '2px';
}

interface BoxShadow {
  heavy: string;
  light: string;
  none: 'none';
  selected: string;
}

interface Colors {
  black: 'hsl(0, 0%, 10%)'; // #1a1a1a
  blue50: 'hsl(191, 100%, 97%)'; // #f0fcff
  blue200: 'hsl(191, 100%, 80%)'; // #99ecff
  blue400: 'hsl(191, 100%, 50%)'; // #00d2ff
  blue600: 'hsl(191, 100%, 40%)'; // #00a7cc
  blue700: 'hsl(191, 100%, 35%)'; // #0092b3
  blue800: 'hsl(191, 100%, 30%)'; // #007d99
  graphThemes: {
    blue: [
      '#c7e9b4',
      '#7ecdbb',
      '#1eb5eb',
      '#1d91bf',
      '#235fa9',
      '#253393',
      '#084181',
    ];
    mixed: [
      '#c7e9b4',
      '#c1d4e7',
      '#7ecdbb',
      '#9fbddb',
      '#1eb5eb',
      '#8d95c6',
      '#1d91bf',
      '#8282ab',
      '#235fa9',
      '#89429e',
      '#253393',
      '#800f7a',
      '#084181',
      '#0b0533',
    ];
    purple: [
      '#c1d4e7',
      '#9fbddb',
      '#8d95c6',
      '#8282ab',
      '#89429e',
      '#800f7a',
      '#0b0533',
    ];
  }; // Used by PrometheusGraph component
  gray50: 'hsl(0, 0%, 96%)'; // #f4f4f4
  gray100: 'hsl(0, 0%, 90%)'; // #e6e6e6
  gray200: 'hsl(0, 0%, 80%)'; // #cccccc

  // #737373
  gray600: 'hsl(0, 0%, 45%)';
  green500: 'hsl(161, 54%, 48%)'; // #38bd93
  orange500: 'hsl(13, 100%, 50%)'; // #ff3700
  orange600: 'hsl(13, 100%, 40%)'; // #cc2c00
  orange700: 'hsl(13, 100%, 35%)'; // #b32700
  // Accent Colors
  // #992100
  orange800: 'hsl(13, 100%, 30%)';
  promQL: {
    attrName: '#00a4db';
    /**
     * GHColors theme by Avi Aryan (http://aviaryan.in)
     * Inspired by Github syntax coloring
     */
    comment: '#bbbbbb';
    deleted: '#9a050f';
    entity: '#36acaa';
    function: '#dc322f';
    metricName: '#2aa198';
    punctuation: '#393a34';
    // Dropdown colors
    salmon: '#ff7c7c';
    string: '#e3116c';
    tag: '#00009f';
  }; // PromQL
  purple25: 'hsl(240, 20%, 98%)'; // #fafafc
  purple50: 'hsl(240, 20%, 95%)'; // #eeeef4
  purple100: 'hsl(240, 20%, 90%)'; // #dfdfea
  purple200: 'hsl(240, 20%, 75%)'; // #b1b1cb
  purple300: 'hsl(240, 20%, 65%)'; // #9494b8
  purple400: 'hsl(240, 20%, 60%)'; // #8585ad

  purple500: 'hsl(240, 20%, 50%)'; // #666699
  purple600: 'hsl(240, 20%, 45%)'; // #5b5b88
  purple700: 'hsl(240, 20%, 35%)'; // #47476b
  purple800: 'hsl(240, 20%, 30%)'; // #3d3d5c
  // Primary Colors
  // #32324b
  purple900: 'hsl(240, 20%, 25%)';

  // Third-party specific colors - not to be used in the theme!
  thirdParty: {
    azure: '#3769bb';
    // Google single-click login
    cornflowerBlue: '#4285f4';
  };
  // #ffffff
  white: 'hsl(0, 0%, 100%)';
  // #d4ab27
  yellow500: 'hsl(46, 69%, 49%)';
}

interface FontFamilies {
  monospace: "'Roboto Mono', monospace";
  regular: "'proxima-nova', Helvetica, Arial, sans-serif";
}

interface FontSizes {
  huge: '48px';
  extraLarge: '32px';
  large: '22px';
  normal: '16px';
  small: '14px';
  tiny: '12px';
}

interface Layers {
  alert: 3;
  dropdown: 5;
  front: 1;
  modal: 7;
  notification: 4;
  toolbar: 2;
  tooltip: 6;
}

type OverlayIconSize = '300px';

interface Spacing {
  base: '16px';
  large: '32px';
  medium: '24px';
  none: '0';
  small: '12px';
  xl: '48px';
  xs: '8px';
  xxl: '64px';
  xxs: '4px';
}

type HexColor = string;

type TextColor = HexColor;

type Unresolved = string;

interface UITheme {
  atoms: Atoms;
  borderRadius: BorderRadius;
  boxShadow: BoxShadow;
  colors: Colors;
  fontFamilies: FontFamilies;
  fontSizes: FontSizes;
  layers: Layers;
  overlayIconSize: OverlayIconSize;
  spacing: Spacing;
  textColor: TextColor;
}

declare module 'weaveworks-ui-components' {
  export interface Theme extends UITheme {}

  export const Button: React.ComponentType<{
    danger?: boolean;
    disabled?: boolean;
    onClick?: (event?: React.MouseEvent<HTMLButtonElement>) => void;
    primary?: boolean;
    selected?: boolean;
    text?: string;
    type?: string;
    title?: string;
  }>;

  export const Dialog: React.ComponentType<{
    actions?: JSX.Element[] | string[];
    active: boolean;
    hideClose?: boolean;
    onClose?: () => void;
    title?: string;
    width?: string;
  }>;

  export interface DropdownItem {
    label: string;
    selectedLabel?: string;
    value: string;
  }

  export const Dropdown: React.ComponentType<{
    disabled?: boolean;
    items: DropdownItem[];
    onChange?: (
      event: React.FormEvent<HTMLInputElement>,
      value: string,
    ) => void;
    placeholder?: string;
    value?: string;
    width?: string;
    withComponent?: () => void;
  }>;

  export const Input: React.ComponentType<{
    autoSelectText?: boolean;
    disabled?: boolean;
    focus?: boolean;
    hideValidationMessage?: boolean;
    inputRef?: () => void;
    label?: string;
    message?: string;
    onChange?: (event: React.ChangeEvent<HTMLInputElement>) => void;
    textarea?: boolean;
    valid?: boolean;
    value?: string;
  }>;

  export const CircularProgress: React.ComponentType<{
    center?: boolean = false;
    inline?: boolean = false;
    size?: 'small' | 'medium' = 'medium';
  }>;

  export const Search: React.ComponentType<{
    onBlur: () => void;
    onChange: (query: string, pinnedTerms: string[]) => void;
    readonly pinnedTerms: string[];
    query: string;
  }>;

  export const DataTable: React.ComponentType<{
    data: {
      [name: string]: any;
    };
    columns: {
      label: string;
      value: string;
    }[];
    children: (t: T) => any[];
  }>;

  export const ResourceDial: React.ComponentType<{
    disabled?: boolean;
    label: string;
    value: number | null; // notice: this differs slightly from the proptype, but is compatible
    to?: string;
  }>;

  export const Alert: React.ComponentType<{
    icon?: string;
    onClose?: () => void;
    title?: string;
    type: 'info' | 'success' | 'warning' | 'error';
    visible?: boolean;
  }>;
}

declare module 'weaveworks-ui-components/lib/theme/selectors' {
  export const spacing = (spacing: keyof Spacing) => '' as ValueOf<Spacing>;
  export const fontSize = (fontSize: keyof FontSizes) =>
    '' as ValueOf<FontSizes>;

  // NOTE: `color` is intentionally left out here since the true color object doesn't actually fit with what styled-components expect since it contains nested objects (e.g. `graphThemes`).
  // export const color = (color: keyof Colors) => '' as ValueOf<Colors>;
  export const borderRadius = (borderRadius: keyof BorderRadius) =>
    '' as ValueOf<BorderRadius>;
  export const boxShadow = (boxShadow: keyof BoxShadow) =>
    '' as ValueOf<BoxShadow>;
}

declare module 'weaveworks-ui-components/lib/theme' {
  interface Theme extends UITheme {}
  const theme: Theme;
  export default theme;
}
