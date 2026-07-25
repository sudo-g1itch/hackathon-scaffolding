// MUI Imports
import Chip from '@mui/material/Chip'
import { useTheme } from '@mui/material/styles'

// Third-party Imports
import PerfectScrollbar from 'react-perfect-scrollbar'

// Type Imports
import type { VerticalMenuContextProps } from '@menu/components/vertical-menu/Menu'

// Component Imports
import { Menu, MenuItem, MenuSection } from '@menu/vertical-menu'

// Config Imports
import { navigationFor } from '@/configs/navigation'

// Hook Imports
import useVerticalNav from '@menu/hooks/useVerticalNav'
import useUnreadMessages from '@/hooks/useUnreadMessages'
import { useAuth } from '@/contexts/AuthContext'

// Styled Component Imports
import StyledVerticalNavExpandIcon from '@menu/styles/vertical/StyledVerticalNavExpandIcon'

// Style Imports
import menuItemStyles from '@core/styles/vertical/menuItemStyles'
import menuSectionStyles from '@core/styles/vertical/menuSectionStyles'

type RenderExpandIconProps = {
  open?: boolean
  transitionDuration?: VerticalMenuContextProps['transitionDuration']
}

type Props = {
  scrollMenu: (container: unknown, isPerfectScrollbar: boolean) => void
}

const RenderExpandIcon = ({ open, transitionDuration }: RenderExpandIconProps) => (
  <StyledVerticalNavExpandIcon open={open} transitionDuration={transitionDuration}>
    <i className='ri-arrow-right-s-line' />
  </StyledVerticalNavExpandIcon>
)

const VerticalMenu = ({ scrollMenu }: Props) => {
  // Hooks
  const theme = useTheme()
  const verticalNavOptions = useVerticalNav()
  const { user } = useAuth()
  const { unread } = useUnreadMessages()

  // Vars
  const { isBreakpointReached, transitionDuration } = verticalNavOptions

  // The menu is built from the shared navigation config rather than hard-coded
  // here, so what a role can see and what a role can open stay one decision.
  const sections = navigationFor(user?.role)

  const ScrollWrapper = isBreakpointReached ? 'div' : PerfectScrollbar

  return (
    // eslint-disable-next-line lines-around-comment
    /* Custom scrollbar instead of browser scroll, remove if you want browser scroll only */
    <ScrollWrapper
      {...(isBreakpointReached
        ? {
            className: 'bs-full overflow-y-auto overflow-x-hidden',
            onScroll: container => scrollMenu(container, false)
          }
        : {
            options: { wheelPropagation: false, suppressScrollX: true },
            onScrollY: container => scrollMenu(container, true)
          })}
    >
      {/* Incase you also want to scroll NavHeader to scroll with Vertical Menu, remove NavHeader from above and paste it below this comment */}
      {/* Vertical Menu */}
      <Menu
        popoutMenuOffset={{ mainAxis: 17 }}
        menuItemStyles={menuItemStyles(verticalNavOptions, theme)}
        renderExpandIcon={({ open }) => <RenderExpandIcon open={open} transitionDuration={transitionDuration} />}
        renderExpandedMenuItemIcon={{ icon: <i className='ri-circle-fill' /> }}
        menuSectionStyles={menuSectionStyles(verticalNavOptions, theme)}
      >

        {sections.map(section => (
          <MenuSection key={section.label} label={section.label}>
            {section.items.map(item => (
              <MenuItem
                key={item.href}
                href={item.href}
                icon={<i className={item.icon} />}
                suffix={
                  item.badge === 'unread' && unread > 0 ? (
                    <Chip size='small' color='error' variant='tonal' label={unread} />
                  ) : undefined
                }
              >
                {item.label}
              </MenuItem>
            ))}
          </MenuSection>
        ))}
      </Menu>
    </ScrollWrapper>
  )
}

export default VerticalMenu
