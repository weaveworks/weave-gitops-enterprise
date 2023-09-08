declare module '*.svg' {
  import React = require('react');
  export const ReactComponent: React.ElementType<React.SVGProps>;
  const src: string;
  export default src;
}
