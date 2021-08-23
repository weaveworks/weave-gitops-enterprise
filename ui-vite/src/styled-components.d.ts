// import original module declarations
import 'styled-components';
import { Theme } from 'weaveworks-ui-components';


declare module 'styled-components' {
  export interface DefaultTheme extends Theme { }
}
