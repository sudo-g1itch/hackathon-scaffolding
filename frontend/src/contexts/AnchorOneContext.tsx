'use client'

// Shared state for the AnchorOne recovery screens: the dashboard snapshot, the
// microphone lifecycle, and one audio element for all spoken playback.
import { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState, type ReactNode } from 'react'

import { anchorOneService } from '@/services/anchorOneService'
import type { Capabilities, Checkin, DashboardData } from '@/types/anchorOneTypes'
import { getApiErrorMessage } from '@/utils/handleApiError'

/** Where the voice check-in currently is, so the UI can narrate each stage. */
export type CheckinPhase = 'idle' | 'recording' | 'processing' | 'done' | 'error'

type AnchorOneContextValue = {
  dashboard: DashboardData | null
  loading: boolean
  error: string | null
  refreshDashboard: () => Promise<void>

  capabilities: Capabilities

  // Voice check-in
  phase: CheckinPhase
  recordingSeconds: number
  result: Checkin | null
  checkinError: string | null
  startRecording: () => Promise<void>

  /** Stops the recorder; analysis continues in its onstop handler. */
  stopRecording: () => void
  cancelRecording: () => void
  submitText: (transcript: string) => Promise<void>
  resetCheckin: () => void

  // Spoken playback (Deepgram TTS)
  speaking: boolean
  speak: (text: string) => Promise<void>
  stopSpeaking: () => void
}

const AnchorOneContext = createContext<AnchorOneContextValue | null>(null)

const DEFAULT_CAPABILITIES: Capabilities = { ai: false, voice: false }

/** Picks a container the browser can actually record, newest codec first. */
const pickMimeType = (): string | undefined => {
  if (typeof MediaRecorder === 'undefined') return undefined

  const candidates = ['audio/webm;codecs=opus', 'audio/webm', 'audio/mp4', 'audio/ogg;codecs=opus']

  return candidates.find(type => MediaRecorder.isTypeSupported(type))
}

const extensionFor = (mimeType: string): string => {
  if (mimeType.includes('mp4')) return 'm4a'
  if (mimeType.includes('ogg')) return 'ogg'

  return 'webm'
}

export const AnchorOneProvider = ({ children }: { children: ReactNode }) => {
  const [dashboard, setDashboard] = useState<DashboardData | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [phase, setPhase] = useState<CheckinPhase>('idle')
  const [recordingSeconds, setRecordingSeconds] = useState(0)
  const [result, setResult] = useState<Checkin | null>(null)
  const [checkinError, setCheckinError] = useState<string | null>(null)

  const [speaking, setSpeaking] = useState(false)

  // Refs, not state: the recorder callbacks fire outside React's render cycle
  // and must always see the live objects.
  const recorderRef = useRef<MediaRecorder | null>(null)
  const streamRef = useRef<MediaStream | null>(null)
  const chunksRef = useRef<Blob[]>([])
  const cancelledRef = useRef(false)
  const audioRef = useRef<HTMLAudioElement | null>(null)
  const audioUrlRef = useRef<string | null>(null)

  const capabilities = dashboard?.capabilities ?? DEFAULT_CAPABILITIES

  const refreshDashboard = useCallback(async () => {
    try {
      setDashboard(await anchorOneService.getDashboard())
      setError(null)
    } catch (err) {
      setError(getApiErrorMessage(err, 'Could not load your dashboard.'))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void refreshDashboard()
  }, [refreshDashboard])

  // Tick the visible recording timer.
  useEffect(() => {
    if (phase !== 'recording') {
      setRecordingSeconds(0)

      return undefined
    }

    const interval = setInterval(() => setRecordingSeconds(seconds => seconds + 1), 1000)

    return () => clearInterval(interval)
  }, [phase])

  const releaseStream = useCallback(() => {
    streamRef.current?.getTracks().forEach(track => track.stop())
    streamRef.current = null
    recorderRef.current = null
    chunksRef.current = []
  }, [])

  const releaseAudioUrl = useCallback(() => {
    if (audioUrlRef.current) {
      URL.revokeObjectURL(audioUrlRef.current)
      audioUrlRef.current = null
    }
  }, [])

  // Never leave the microphone open or a blob URL leaked on unmount.
  useEffect(
    () => () => {
      releaseStream()
      releaseAudioUrl()
      audioRef.current?.pause()
    },
    [releaseStream, releaseAudioUrl]
  )

  const analyze = useCallback(
    async (run: () => Promise<Checkin>) => {
      setPhase('processing')
      setCheckinError(null)

      try {
        const checkin = await run()

        setResult(checkin)
        setPhase('done')
        await refreshDashboard()
      } catch (err) {
        setCheckinError(getApiErrorMessage(err, 'We could not analyse that check-in.'))
        setPhase('error')
      }
    },
    [refreshDashboard]
  )

  const startRecording = useCallback(async () => {
    setCheckinError(null)
    setResult(null)
    cancelledRef.current = false

    if (typeof navigator === 'undefined' || !navigator.mediaDevices?.getUserMedia) {
      setCheckinError('This browser cannot record audio. Type your check-in instead.')
      setPhase('error')

      return
    }

    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
      const mimeType = pickMimeType()
      const recorder = new MediaRecorder(stream, mimeType ? { mimeType } : undefined)

      chunksRef.current = []

      recorder.ondataavailable = event => {
        if (event.data.size > 0) chunksRef.current.push(event.data)
      }

      recorder.onstop = () => {
        const type = recorder.mimeType || mimeType || 'audio/webm'
        const audio = new Blob(chunksRef.current, { type })

        releaseStream()

        if (cancelledRef.current) {
          setPhase('idle')

          return
        }

        if (audio.size === 0) {
          setCheckinError('That recording came through empty. Please try again.')
          setPhase('error')

          return
        }

        void analyze(() => anchorOneService.checkin(audio, `checkin.${extensionFor(type)}`))
      }

      streamRef.current = stream
      recorderRef.current = recorder
      recorder.start()
      setPhase('recording')
    } catch {
      // Overwhelmingly this is the user denying the permission prompt.
      setCheckinError('Microphone access was blocked. Allow it in your browser, or type your check-in instead.')
      setPhase('error')
    }
  }, [analyze, releaseStream])

  const stopRecording = useCallback(() => {
    cancelledRef.current = false

    if (recorderRef.current?.state === 'recording') {
      recorderRef.current.stop()
    }
  }, [])

  const cancelRecording = useCallback(() => {
    cancelledRef.current = true

    if (recorderRef.current?.state === 'recording') {
      recorderRef.current.stop()
    } else {
      releaseStream()
      setPhase('idle')
    }
  }, [releaseStream])

  const submitText = useCallback(
    async (transcript: string) => {
      await analyze(() => anchorOneService.analyzeText(transcript))
    },
    [analyze]
  )

  const resetCheckin = useCallback(() => {
    setResult(null)
    setCheckinError(null)
    setPhase('idle')
  }, [])

  const stopSpeaking = useCallback(() => {
    audioRef.current?.pause()
    setSpeaking(false)
  }, [])

  const speak = useCallback(
    async (text: string) => {
      try {
        stopSpeaking()
        releaseAudioUrl()

        const blob = await anchorOneService.speak(text)
        const url = URL.createObjectURL(blob)

        audioUrlRef.current = url

        const audio = audioRef.current ?? new Audio()

        audioRef.current = audio
        audio.onended = () => setSpeaking(false)
        audio.onerror = () => setSpeaking(false)
        audio.src = url
        setSpeaking(true)
        await audio.play()
      } catch (err) {
        setSpeaking(false)
        setError(getApiErrorMessage(err, 'Could not play that audio.'))
      }
    },
    [releaseAudioUrl, stopSpeaking]
  )

  const value = useMemo<AnchorOneContextValue>(
    () => ({
      dashboard,
      loading,
      error,
      refreshDashboard,
      capabilities,
      phase,
      recordingSeconds,
      result,
      checkinError,
      startRecording,
      stopRecording,
      cancelRecording,
      submitText,
      resetCheckin,
      speaking,
      speak,
      stopSpeaking
    }),
    [
      dashboard,
      loading,
      error,
      refreshDashboard,
      capabilities,
      phase,
      recordingSeconds,
      result,
      checkinError,
      startRecording,
      stopRecording,
      cancelRecording,
      submitText,
      resetCheckin,
      speaking,
      speak,
      stopSpeaking
    ]
  )

  return <AnchorOneContext.Provider value={value}>{children}</AnchorOneContext.Provider>
}

export const useAnchorOne = () => {
  const context = useContext(AnchorOneContext)

  if (!context) {
    throw new Error('useAnchorOne must be used within an AnchorOneProvider')
  }

  return context
}
