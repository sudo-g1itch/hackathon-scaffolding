// AnchorOne recovery domain types.
//
// These mirror the DTOs in backend/internal/service/recoverai_service.go one
// for one. When a field changes there, change it here — the two files are the
// API contract and nothing else should redefine these shapes.

export type RiskLevel = 'LOW' | 'MEDIUM' | 'HIGH'

export type CheckinSource = 'voice' | 'text'

export type CoachRole = 'user' | 'ai'

export type TimelineEventType = 'checkin' | 'emergency'

/** Which optional integrations the server actually has keys for. */
export type Capabilities = {
  ai: boolean
  voice: boolean
}

export type Checkin = {
  id: string
  created_at: string
  user_id: string
  transcript: string
  summary: string
  emotion: string
  craving: number
  risk: RiskLevel
  triggers: string[] | null
  recommended_actions: string[] | null
  source: CheckinSource
}

export type CoachMessage = {
  id: string
  created_at: string
  role: CoachRole
  message: string
}

export type EmergencyLog = {
  id: string
  created_at: string
  actions: string[] | null
  generated_script: string
  grounding_exercise: string
  encouraging_message: string
}

export type EmergencyPlan = {
  immediate_actions: string[]
  emergency_sms: string
  grounding_exercise: string
  encouraging_message: string
}

export type EmergencyResult = {
  log: EmergencyLog
  plan: EmergencyPlan
}

export type RecoveryProfile = {
  goal: string
  substance: string
  caregiver_id: string | null
  caregiver_name: string
  caregiver_phone: string
  emergency_contact: string

  /** Full name of the linked caregiver account, when one is linked. */
  linked_caregiver_name: string

  /**
   * The user's consent for their caregiver to read what a check-in SAID, not
   * just how risky it scored. Off by default; only the user can turn it on.
   */
  share_checkin_details: boolean
}

export type ProfileInput = {
  goal: string
  substance: string
  caregiver_name: string
  caregiver_phone: string
  emergency_contact: string
  share_checkin_details: boolean
}

export type DashboardData = {
  current_mood: string
  risk_badge: RiskLevel
  craving_level: number
  recovery_streak: number
  total_checkins: number
  emergency_count: number
  last_checkin: Checkin | null
  profile: RecoveryProfile | null
  capabilities: Capabilities
  goals: GoalSummary
  unread_messages: number
}

/** One entry of the merged, reverse-chronological recovery history. */
export type TimelineEvent = {
  id: string
  type: TimelineEventType
  occurred_at: string
  risk?: RiskLevel
  emotion?: string
  craving?: number
  summary?: string
  triggers?: string[]
  actions?: string[]
  generated_script?: string
  grounding_exercise?: string
  encouraging_message?: string
  source?: CheckinSource
}

/**
 * A person a caregiver supports, as shown in the caregiver's list.
 *
 * Deliberately carries no transcript, summary or triggers — the backend does
 * not send them here, because private conversations are never shared.
 */
export type CaregiverPatient = {
  user_id: string
  name: string
  goal: string
  substance: string
  risk: RiskLevel
  emotion: string
  craving: number
  recovery_streak: number
  last_checkin_at: string | null
  emergency_count: number
  active_goals: number
  completed_goals: number
  average_progress: number
  unread_messages: number
}

export type CaregiverOption = {
  id: string
  name: string
}

// --- recovery plan (goals) ---------------------------------------------------

export type GoalStatus = 'active' | 'completed' | 'paused' | 'archived'

export type GoalCategory = 'sobriety' | 'health' | 'routine' | 'social' | 'work' | 'other'

/** Who put a goal on the plan, or who wrote an entry in its progress log. */
export type AuthorRole = 'user' | 'caregiver'

export type GoalUpdateKind = 'progress' | 'note' | 'encouragement' | 'status'

export type Goal = {
  id: string
  user_id: string
  title: string
  description: string
  category: GoalCategory
  status: GoalStatus
  target_value: number
  current_value: number
  unit: string

  /** Computed server-side so every screen agrees on how far along a goal is. */
  progress_percent: number
  target_date: string | null
  completed_at: string | null
  created_by_role: AuthorRole
  created_at: string
  updated_at: string

  /** Null when the goal has no deadline; negative once a deadline has passed. */
  days_remaining: number | null
}

export type GoalUpdate = {
  id: string
  goal_id: string
  author_id: string
  author_name: string
  author_role: AuthorRole
  kind: GoalUpdateKind
  value: number
  delta: number
  note: string
  created_at: string
}

export type GoalDetail = {
  goal: Goal
  updates: GoalUpdate[]
}

export type GoalSummary = {
  active: number
  completed: number
  paused: number
  archived: number
  total: number
  average_progress: number
  next_goal_title: string
}

export type GoalInput = {
  title: string
  description: string
  category: GoalCategory
  target_value: number
  unit: string
  target_date: string | null
}

/** Partial update — an omitted key leaves that field alone. */
export type GoalPatch = Partial<GoalInput> & {
  status?: GoalStatus
  current_value?: number

  /** Removes a deadline; without it there is no way to express "clear". */
  clear_target_date?: boolean
}

export type ProgressInput = {
  delta?: number
  value?: number
  note?: string
  kind?: GoalUpdateKind
}

// --- caregiver <-> user messaging --------------------------------------------

export type SupportMessage = {
  id: string
  patient_id: string
  caregiver_id: string
  sender_id: string
  sender_role: AuthorRole
  body: string
  read_at: string | null
  created_at: string
}

/**
 * The shared conversation. `linked` is false when the user has not chosen a
 * caregiver yet — an ordinary state the UI answers with a prompt, not an error.
 */
export type SupportThread = {
  linked: boolean
  patient_id: string
  patient_name: string
  caregiver_id: string | null
  caregiver_name: string
  messages: SupportMessage[]
  unread: number
}

/**
 * One check-in as somebody other than its author may see it.
 *
 * `summary` and `triggers` are present only when the person in recovery has
 * switched on `share_checkin_details`. The raw transcript is never sent.
 */
export type PatientCheckin = {
  id: string
  occurred_at: string
  risk: RiskLevel
  emotion: string
  craving: number
  source: CheckinSource
  summary?: string
  triggers?: string[]
}

/** How the viewer is related to the person whose record they are reading. */
export type CareRelation = 'self' | 'caregiver' | 'admin'

export type PatientOverview = {
  patient: CaregiverPatient
  checkins: PatientCheckin[]
  goals: Goal[]
  goal_summary: GoalSummary
  shares_checkin_details: boolean
  relation: CareRelation
}

export type RiskColor = 'success' | 'warning' | 'error'

/** Shared risk → MUI palette mapping, so every screen colours risk the same. */
export const RISK_COLORS: Record<RiskLevel, RiskColor> = {
  LOW: 'success',
  MEDIUM: 'warning',
  HIGH: 'error'
}

export const resolveRiskColor = (risk?: string | null): RiskColor =>
  RISK_COLORS[(risk ?? 'LOW') as RiskLevel] ?? 'success'

/** Numeric encoding of risk for the trend chart's Y axis. */
export const RISK_SCORES: Record<RiskLevel, number> = {
  LOW: 1,
  MEDIUM: 2,
  HIGH: 3
}

/**
 * Presentation for a goal category — one definition, so the goal card, the
 * create dialog and the caregiver's view pick the same icon and colour.
 */
export const GOAL_CATEGORIES: Record<GoalCategory, { label: string; icon: string; color: RiskColor | 'primary' | 'info' }> =
  {
    sobriety: { label: 'Sobriety', icon: 'ri-shield-check-line', color: 'primary' },
    health: { label: 'Health', icon: 'ri-heart-pulse-line', color: 'success' },
    routine: { label: 'Daily routine', icon: 'ri-calendar-check-line', color: 'info' },
    social: { label: 'Relationships', icon: 'ri-group-line', color: 'warning' },
    work: { label: 'Work & study', icon: 'ri-briefcase-line', color: 'info' },
    other: { label: 'Other', icon: 'ri-flag-line', color: 'primary' }
  }

/** Ordered for the category picker. */
export const GOAL_CATEGORY_ORDER: GoalCategory[] = ['sobriety', 'health', 'routine', 'social', 'work', 'other']

export const GOAL_STATUSES: Record<GoalStatus, { label: string; color: 'success' | 'warning' | 'secondary' | 'info' }> = {
  active: { label: 'Active', color: 'info' },
  completed: { label: 'Completed', color: 'success' },
  paused: { label: 'Paused', color: 'warning' },
  archived: { label: 'Archived', color: 'secondary' }
}

/** Units a goal can be measured in. Free-form on the wire; these are the presets. */
export const GOAL_UNITS = ['days', 'weeks', 'sessions', 'meetings', 'times', 'steps'] as const

export const resolveGoalCategory = (category?: string | null) =>
  GOAL_CATEGORIES[(category ?? 'other') as GoalCategory] ?? GOAL_CATEGORIES.other

export const resolveGoalStatus = (status?: string | null) =>
  GOAL_STATUSES[(status ?? 'active') as GoalStatus] ?? GOAL_STATUSES.active
