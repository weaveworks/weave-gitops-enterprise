import { Flex, useFeatureFlags } from '@weaveworks/weave-gitops';
import { useLocation } from 'react-router-dom';

import { Tooltip } from '../Shared';
import { getNavItems, NavigationItem } from './NavItemList';
import { NavGroupItemWrapper, NavItem } from './StyledComponent';

const NavLinkItem = ({
  item,
  className,
  collapsed = false,
}: {
  item: NavigationItem;
  className: string;
  collapsed: boolean;
}) => {
  return (
    <NavItem to={item.link} className={`route-nav ${className}`}>
      {!collapsed ? (
        <Tooltip arrow placement="right" title={item.name}>
          <Flex center>{item.icon}</Flex>
        </Tooltip>
      ) : (
        <Flex center>{item.icon}</Flex>
      )}
      <span className="toggleOpacity ellipsis route-item">{item.name}</span>
    </NavItem>
  );
};

const NavItems = ({ collapsed = false }: { collapsed: boolean }) => {
  const { data: flagsRes } = useFeatureFlags();
  const location = useLocation();
  const groupItems = getNavItems(flagsRes);
  return (
    <>
      {groupItems.map(({ text, items }) => {
        return (
          <NavGroupItemWrapper column key={text} collapsed={collapsed}>
            <div className="title toggleOpacity ellipsis">{text}</div>
            {items.map(item => (
              <NavLinkItem
                key={item.name}
                collapsed={collapsed}
                item={item}
                className={
                  item.relatedRoutes?.some(link =>
                    location.pathname.includes(link),
                  )
                    ? 'nav-link-active'
                    : ''
                }
              />
            ))}
          </NavGroupItemWrapper>
        );
      })}
    </>
  );
};
export default NavItems;
