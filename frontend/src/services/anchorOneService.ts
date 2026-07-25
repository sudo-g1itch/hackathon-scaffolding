// Single HTTP surface for the AnchorOne recovery features.
//
// Paths are relative: the axios instance already carries `/api/v1` in its
// baseURL, so prefixing it here would resolve to /api/v1/api/v1/... and 404.
//
// Every call unwraps the standard envelope and returns typed data, so screens
// never touch `response.data.data` or `any`.
import axios from '@/libs/axios'

import type { StandardResponse } from '@/types/apiTypes'
import type {
  Capabilities,
  CaregiverOption,
  CaregiverPatient,
  Checkin,
  CoachMessage,
  DashboardData,
  EmergencyAlert,
  EmergencyAlertInput,
  EmergencyResult,
  Goal,
  GoalDetail,
  GoalInput,
  GoalPatch,
  GoalSummary,
  PatientOverview,
  ProfileInput,
  ProgressInput,
  RecoveryProfile,
  SupportThread,
  TimelineEvent
} from '@/types/anchorOneTypes'

const AUDIO_FIELD = 'audio'

/** Unwraps the success envelope, failing loudly if the payload is missing. */
const unwrap = <T>(response: { data: StandardResponse<T> }): T => {
  const payload = response.data.data

  if (payload === undefined || payload === null) {
    throw new Error('The server returned an empty response.')
  }

  return payload
}

/** Builds the multipart body a check-in / transcription upload expects. */
const audioForm = (audio: Blob, filename: string): FormData => {
  const form = new FormData()

  form.append(AUDIO_FIELD, audio, filename)

  return form
}

export const anchorOneService = {
  /** Which AI integrations this server has configured. */
  getCapabilities: async (): Promise<Capabilities> =>
    unwrap(await axios.get<StandardResponse<Capabilities>>('/capabilities')),

  /** Voice check-in: upload audio, get back the analysed check-in. */
  checkin: async (audio: Blob, filename = 'checkin.webm'): Promise<Checkin> =>
    unwrap(
      await axios.post<StandardResponse<Checkin>>('/checkin', audioForm(audio, filename), {
        headers: { 'Content-Type': 'multipart/form-data' }
      })
    ),

  /** Typed check-in — same analysis, for when speaking is not an option. */
  analyzeText: async (transcript: string): Promise<Checkin> =>
    unwrap(await axios.post<StandardResponse<Checkin>>('/risk', { transcript })),

  /** Speech to text only — no analysis, nothing stored. */
  transcribe: async (audio: Blob, filename = 'audio.webm'): Promise<string> => {
    const data = unwrap(
      await axios.post<StandardResponse<{ transcript: string }>>('/voice/transcribe', audioForm(audio, filename), {
        headers: { 'Content-Type': 'multipart/form-data' }
      })
    )

    return data.transcript
  },

  /**
   * Text to speech. Returns raw MP3 bytes, so this call bypasses the JSON
   * envelope and asks axios for a blob.
   */
  speak: async (text: string): Promise<Blob> => {
    const response = await axios.post<Blob>('/voice/speak', { text }, { responseType: 'blob' })

    return response.data
  },

  getDashboard: async (): Promise<DashboardData> =>
    unwrap(await axios.get<StandardResponse<DashboardData>>('/dashboard')),

  getTimeline: async (): Promise<TimelineEvent[]> =>
    unwrap(await axios.get<StandardResponse<TimelineEvent[]>>('/timeline')),

  triggerEmergency: async (): Promise<EmergencyResult> =>
    unwrap(await axios.post<StandardResponse<EmergencyResult>>('/emergency')),

  /** Attaches a voice note to an alert. Only its transcript is stored. */
  attachEmergencyNote: async (logId: string, audio: Blob, filename = 'note.webm'): Promise<EmergencyResult> =>
    unwrap(
      await axios.post<StandardResponse<EmergencyResult>>(`/emergency/${logId}/note`, audioForm(audio, filename), {
        headers: { 'Content-Type': 'multipart/form-data' }
      })
    ),

  /** Sends the chosen script to the linked caregiver. This one really sends. */
  sendEmergencyAlert: async (logId: string, input: EmergencyAlertInput): Promise<EmergencyResult> =>
    unwrap(await axios.post<StandardResponse<EmergencyResult>>(`/emergency/${logId}/alert`, input)),

  /** Caregiver confirming they have seen an alert. */
  acknowledgeEmergency: async (logId: string): Promise<EmergencyAlert> =>
    unwrap(await axios.post<StandardResponse<EmergencyAlert>>(`/emergency/${logId}/acknowledge`)),

  sendCoachMessage: async (message: string): Promise<CoachMessage[]> =>
    unwrap(await axios.post<StandardResponse<CoachMessage[]>>('/coach/chat', { message })),

  getCoachHistory: async (): Promise<CoachMessage[]> =>
    unwrap(await axios.get<StandardResponse<CoachMessage[]>>('/coach/history')),

  educate: async (query: string): Promise<string> => {
    const data = unwrap(await axios.post<StandardResponse<{ result: string }>>('/education', { query }))

    return data.result
  },

  getProfile: async (): Promise<RecoveryProfile> =>
    unwrap(await axios.get<StandardResponse<RecoveryProfile>>('/profile')),

  updateProfile: async (input: ProfileInput): Promise<RecoveryProfile> =>
    unwrap(await axios.put<StandardResponse<RecoveryProfile>>('/profile', input)),

  getCaregiverData: async (): Promise<CaregiverPatient[]> =>
    unwrap(await axios.get<StandardResponse<CaregiverPatient[]>>('/caregiver')),

  getAvailableCaregivers: async (): Promise<CaregiverOption[]> =>
    unwrap(await axios.get<StandardResponse<CaregiverOption[]>>('/caregivers')),

  /** Links a caregiver, or unlinks the current one when passed null. */
  setCaregiver: async (caregiverId: string | null): Promise<RecoveryProfile> =>
    unwrap(await axios.put<StandardResponse<RecoveryProfile>>('/profile/caregiver', { caregiver_id: caregiverId })),

  // --- recovery plan (goals) -------------------------------------------------
  //
  // The bare /goals paths always act on the signed-in user's own plan. To read
  // or add to somebody else's, use the patient-scoped calls below; the server
  // decides whether the caller is allowed to.

  getGoals: async (): Promise<Goal[]> => unwrap(await axios.get<StandardResponse<Goal[]>>('/goals')),

  getGoalSummary: async (): Promise<GoalSummary> =>
    unwrap(await axios.get<StandardResponse<GoalSummary>>('/goals/summary')),

  getGoal: async (goalId: string): Promise<GoalDetail> =>
    unwrap(await axios.get<StandardResponse<GoalDetail>>(`/goals/${goalId}`)),

  createGoal: async (input: GoalInput): Promise<Goal> =>
    unwrap(await axios.post<StandardResponse<Goal>>('/goals', input)),

  updateGoal: async (goalId: string, patch: GoalPatch): Promise<Goal> =>
    unwrap(await axios.put<StandardResponse<Goal>>(`/goals/${goalId}`, patch)),

  deleteGoal: async (goalId: string): Promise<void> => {
    await axios.delete(`/goals/${goalId}`)
  },

  /** Logs movement, a note, or (from a caregiver) a word of encouragement. */
  logGoalProgress: async (goalId: string, input: ProgressInput): Promise<GoalDetail> =>
    unwrap(await axios.post<StandardResponse<GoalDetail>>(`/goals/${goalId}/progress`, input)),

  // --- patient-scoped (caregiver) views --------------------------------------

  getPatientOverview: async (patientId: string): Promise<PatientOverview> =>
    unwrap(await axios.get<StandardResponse<PatientOverview>>(`/patients/${patientId}`)),

  getPatientGoals: async (patientId: string): Promise<Goal[]> =>
    unwrap(await axios.get<StandardResponse<Goal[]>>(`/patients/${patientId}/goals`)),

  /** A caregiver suggesting a goal for someone they support. */
  createPatientGoal: async (patientId: string, input: GoalInput): Promise<Goal> =>
    unwrap(await axios.post<StandardResponse<Goal>>(`/patients/${patientId}/goals`, input)),

  // --- messaging -------------------------------------------------------------
  //
  // One thread per (patient, caregiver) pair, read and written by both sides.
  // A user passes their own id; a caregiver passes the patient's.

  getSupportThread: async (patientId: string): Promise<SupportThread> =>
    unwrap(await axios.get<StandardResponse<SupportThread>>(`/patients/${patientId}/messages`)),

  sendSupportMessage: async (patientId: string, body: string): Promise<SupportThread> =>
    unwrap(await axios.post<StandardResponse<SupportThread>>(`/patients/${patientId}/messages`, { body })),

  markThreadRead: async (patientId: string): Promise<void> => {
    await axios.post(`/patients/${patientId}/messages/read`)
  },

  getUnreadCount: async (): Promise<number> => {
    const data = unwrap(await axios.get<StandardResponse<{ unread: number }>>('/messages/unread'))

    return data.unread
  }
}
