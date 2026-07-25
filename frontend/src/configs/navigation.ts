// The single source of truth for what each role may reach.
//
// Both the sidebar and the route guard read this file, so a screen can never be
// hidden from the menu while still being reachable by typing its URL — or the
// reverse. The backend enforces the same boundaries independently; this exists
// so the UI never offers somebody a door that will be shut in their face.
import type { Role } from '@/types/apiTypes'

/**
 * Roles that experience AnchorOne as a person in recovery. `manager` is not an
 * AnchorOne domain role, so it falls back to the ordinary user experience
 * rather than being locked out of the product entirely.
 */
export const RECOVERY_ROLES: Role[] = ['user', 'manager']

/** Roles that support someone else. Admins are included for oversight. */
export const CARE_ROLES: Role[] = ['caregiver', 'admin']

export const ADMIN_ROLES: Role[] = ['admin']

export const ALL_ROLES: Role[] = ['admin', 'manager', 'user', 'caregiver']

export type NavItem = {
  label: string
  href: string
  icon: string
  roles: Role[]

  /** Show the unread-message count next to this item. */
  badge?: 'unread'
}

export type NavSection = {
  label: string
  items: NavItem[]
}

export const NAV_SECTIONS: NavSection[] = [
  {
    label: 'My Recovery',
    items: [
      {
        label: 'Dashboard',
        href: '/anchor-one/dashboard',
        icon: 'ri-dashboard-line',
        roles: RECOVERY_ROLES
      },
      { label: 'AI Coach', href: '/anchor-one/coach', icon: 'ri-robot-line', roles: RECOVERY_ROLES },
      { label: 'My Goals', href: '/anchor-one/goals', icon: 'ri-flag-line', roles: RECOVERY_ROLES },
      { label: 'Timeline', href: '/anchor-one/timeline', icon: 'ri-history-line', roles: RECOVERY_ROLES },
      {
        label: 'My Caregiver',
        href: '/anchor-one/messages',
        icon: 'ri-chat-heart-line',
        roles: RECOVERY_ROLES,
        badge: 'unread'
      },
      {
        label: 'My Recovery Plan',
        href: '/anchor-one/profile',
        icon: 'ri-user-heart-line',
        roles: RECOVERY_ROLES
      }
    ]
  },
  {
    label: 'Care Team',
    items: [
      {
        label: 'People I Support',
        href: '/anchor-one/caregiver',
        icon: 'ri-group-line',
        roles: CARE_ROLES,
        badge: 'unread'
      }
    ]
  },
  {
    label: 'Learn',
    items: [{ label: 'Education', href: '/anchor-one/education', icon: 'ri-book-read-line', roles: ALL_ROLES }]
  },
  {
    label: 'Administration',
    items: [
      { label: 'User Management', href: '/admin/users', icon: 'ri-user-line', roles: ADMIN_ROLES },
      { label: 'Roles', href: '/admin/roles', icon: 'ri-shield-user-line', roles: ADMIN_ROLES },
      { label: 'Permissions', href: '/admin/permissions', icon: 'ri-lock-2-line', roles: ADMIN_ROLES }
    ]
  }
]

/**
 * Routes that are reachable but never listed in the menu — detail screens you
 * arrive at from a list. They still need an access rule, which is exactly why
 * this lives here rather than being implied by NAV_SECTIONS.
 */
const UNLISTED_ROUTES: { href: string; roles: Role[] }[] = [
  // A caregiver's detail view of one person.
  { href: '/anchor-one/caregiver/', roles: CARE_ROLES }
]

/** Every access rule, longest path first so the most specific one wins. */
const ACCESS_RULES: { href: string; roles: Role[] }[] = [
  ...NAV_SECTIONS.flatMap(section => section.items.map(item => ({ href: item.href, roles: item.roles }))),
  ...UNLISTED_ROUTES
].sort((a, b) => b.href.length - a.href.length)

/**
 * The roles allowed on a path, or null when the path has no rule — an
 * unrecognised route is left to Next.js's not-found rather than being guessed at.
 */
export const rolesForPath = (pathname: string): Role[] | null =>
  ACCESS_RULES.find(rule => pathname === rule.href || pathname.startsWith(rule.href))?.roles ?? null

/** Whether a role may open a path. Unknown paths are permitted. */
export const canAccessPath = (pathname: string, role: Role | null | undefined): boolean => {
  const roles = rolesForPath(pathname)

  if (!roles) return true
  if (!role) return false

  return roles.includes(role)
}

/** The menu, filtered to one role, with empty sections dropped. */
export const navigationFor = (role: Role | null | undefined): NavSection[] => {
  if (!role) return []

  return NAV_SECTIONS.map(section => ({
    ...section,
    items: section.items.filter(item => item.roles.includes(role))
  })).filter(section => section.items.length > 0)
}

/**
 * Where a role should land after signing in: the first screen they can
 * actually use, rather than a dashboard that would immediately deny them.
 */
export const landingPathFor = (role: Role | null | undefined): string => {
  const [firstSection] = navigationFor(role)
  const firstItem = firstSection?.items[0]

  return firstItem?.href ?? '/anchor-one/education'
}
