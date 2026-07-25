'use client'

// Reads text aloud through Deepgram TTS. Used wherever the PRD asks for spoken
// output — the emergency plan, the coach's reply, an education answer.
import Button from '@mui/material/Button'
import CircularProgress from '@mui/material/CircularProgress'
import IconButton from '@mui/material/IconButton'
import Tooltip from '@mui/material/Tooltip'

import { useAnchorOne } from '@/contexts/AnchorOneContext'

type SpeakButtonProps = {
  text: string

  /** Icon-only rendering, for dense rows like a chat bubble. */
  iconOnly?: boolean
  label?: string
  size?: 'small' | 'medium' | 'large'
}

const SpeakButton = ({ text, iconOnly = false, label = 'Read aloud', size = 'small' }: SpeakButtonProps) => {
  const { speak, stopSpeaking, speaking, capabilities } = useAnchorOne()

  const disabled = !capabilities.voice || !text.trim()

  const tooltip = capabilities.voice ? label : 'Voice playback is not configured on this server'

  const handleClick = () => {
    if (speaking) {
      stopSpeaking()

      return
    }

    void speak(text)
  }

  if (iconOnly) {
    return (
      <Tooltip title={tooltip}>
        <span>
          <IconButton size={size === 'large' ? 'medium' : 'small'} onClick={handleClick} disabled={disabled}>
            <i className={speaking ? 'ri-stop-circle-line' : 'ri-volume-up-line'} />
          </IconButton>
        </span>
      </Tooltip>
    )
  }

  return (
    <Tooltip title={tooltip}>
      <span>
        <Button
          size={size}
          variant='outlined'
          color='secondary'
          onClick={handleClick}
          disabled={disabled}
          startIcon={
            speaking ? (
              <CircularProgress size={16} color='inherit' />
            ) : (
              <i className='ri-volume-up-line' />
            )
          }
        >
          {speaking ? 'Stop' : label}
        </Button>
      </span>
    </Tooltip>
  )
}

export default SpeakButton
