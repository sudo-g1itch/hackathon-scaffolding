// The single source of truth for lifecycle-status → semantic color across the
// app. Screens render statuses through <StatusChip>, which resolves here.
import type { ThemeColor } from '@core/types'

export const STATUS_COLORS: Record<string, ThemeColor> = {
  // Terminal / good
  Active: 'success',
  Paid: 'success',
  Approved: 'success',
  Processed: 'success',
  Eligible: 'success',
  Closed: 'success',
  Invoiced: 'success',
  Completed: 'success',
  Finalized: 'success',
  Ready: 'success',

  // In-flight / informational
  Submitted: 'info',
  Posted: 'info',
  Issued: 'info',
  In_Progress: 'info',
  Partially_Paid: 'info',
  For_Billing: 'info',

  // Needs attention
  Pending: 'warning',
  Resubmitted: 'warning',
  PartiallyPaid: 'warning',
  Pending_Upload: 'warning',
  For_Review: 'warning',

  // Failed / blocked
  Inactive: 'error',
  Rejected: 'error',
  Cancelled: 'error',
  Error: 'error',
  Failed: 'error',
  Ineligible: 'error',

  // Not started
  Open: 'secondary',
  New: 'secondary',
  Draft: 'secondary',

  // Awaiting a decision
  Pending_For_Approval: 'primary'
}

export const resolveStatusColor = (status: string, overrides?: Record<string, ThemeColor>): ThemeColor =>
  overrides?.[status] ?? STATUS_COLORS[status] ?? 'secondary'

export const FAULT_COLORS: Record<string, ThemeColor> = {
  ProviderData: 'warning',
  Coding: 'info',
  Payer: 'error',
  Mixed: 'secondary',
  Unknown: 'secondary'
}

export const resolveFaultColor = (fault?: string | null): ThemeColor =>
  (fault && FAULT_COLORS[fault]) || 'secondary'

export const formatStatus = (status?: string | null): string => {
  if (!status) return ''

  return status
    .replace(/_/g, ' ')
    .replace(/([a-z0-9])([A-Z])/g, '$1 $2')
    .replace(/\s+/g, ' ')
    .trim()
}

export const resolveStatusLabel = (
  t: ((key: string) => string) | undefined,
  status: string | undefined | null,
  namespace: string
): string => {
  if (!status) return ''

  if (!t) return formatStatus(status)

  const key = `${namespace}.Status${status.replace(/_/g, '')}`
  const translated = t(key)

  return translated === key ? formatStatus(status) : translated
}
