'use client'

// AI Education Assistant — answers recovery questions on demand instead of
// serving static articles.
import { useState } from 'react'

import Alert from '@mui/material/Alert'
import Box from '@mui/material/Box'
import Button from '@mui/material/Button'
import Chip from '@mui/material/Chip'
import CircularProgress from '@mui/material/CircularProgress'
import Paper from '@mui/material/Paper'
import TextField from '@mui/material/TextField'
import Typography from '@mui/material/Typography'
import ReactMarkdown from 'react-markdown'

import SpeakButton from '@components/anchor-one/SpeakButton'
import { useAnchorOne } from '@/contexts/AnchorOneContext'
import { anchorOneService } from '@/services/anchorOneService'
import { getApiErrorMessage } from '@/utils/handleApiError'

const SUGGESTED = [
  'Why do cravings happen?',
  'What is the 5-4-3-2-1 grounding method?',
  'What does relapse actually mean?',
  'How long do withdrawal symptoms last?',
  'How do I handle a trigger I cannot avoid?'
]

const EducationAssistant = () => {
  const { capabilities } = useAnchorOne()

  const [query, setQuery] = useState('')
  const [asked, setAsked] = useState<string | null>(null)
  const [result, setResult] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const ask = async (text: string) => {
    const question = text.trim()

    if (!question || loading) return

    setLoading(true)
    setError(null)
    setResult(null)
    setAsked(question)
    setQuery(question)

    try {
      setResult(await anchorOneService.educate(question))
    } catch (err) {
      setError(getApiErrorMessage(err, 'Could not load an answer just now.'))
    } finally {
      setLoading(false)
    }
  }

  return (
    <Box className='flex flex-col gap-6'>
      <Box>
        <Typography variant='h4' fontWeight={700}>
          Education Assistant
        </Typography>
        <Typography variant='body2' color='text.secondary'>
          Ask anything about recovery, cravings or substance use — answered in plain language.
        </Typography>
      </Box>

      {!capabilities.ai && <Alert severity='warning'>The education assistant is not configured on this server.</Alert>}

      <Box className='flex flex-wrap gap-2'>
        {SUGGESTED.map(suggestion => (
          <Chip
            key={suggestion}
            label={suggestion}
            variant='tonal'
            color='info'
            onClick={() => void ask(suggestion)}
            disabled={loading || !capabilities.ai}
          />
        ))}
      </Box>

      <Box className='flex gap-3'>
        <TextField
          fullWidth
          placeholder='e.g. Why do cravings happen?'
          value={query}
          onChange={event => setQuery(event.target.value)}
          onKeyDown={event => {
            if (event.key === 'Enter') {
              event.preventDefault()
              void ask(query)
            }
          }}
          disabled={loading || !capabilities.ai}
        />
        <Button
          variant='contained'
          onClick={() => void ask(query)}
          disabled={loading || !query.trim() || !capabilities.ai}
          sx={{ minWidth: 120 }}
          startIcon={loading ? undefined : <i className='ri-search-line' />}
        >
          {loading ? <CircularProgress size={22} color='inherit' /> : 'Ask'}
        </Button>
      </Box>

      {error && <Alert severity='error'>{error}</Alert>}

      {result && (
        <Paper variant='outlined' sx={{ p: 5, borderRadius: 2 }}>
          <Box className='flex items-start justify-between gap-4 mbe-3'>
            <Typography variant='h6'>{asked}</Typography>
            <SpeakButton text={result} label='Read aloud' />
          </Box>
          <Box
            sx={{
              '& p': { marginBlock: '0.5rem' },
              '& ul, & ol': { paddingInlineStart: '1.5rem' },
              '& li': { marginBlock: '0.25rem' },
              '& h1, & h2, & h3': { fontSize: '1.05rem', fontWeight: 600, marginBlockStart: '1rem' }
            }}
          >
            <ReactMarkdown>{result}</ReactMarkdown>
          </Box>
        </Paper>
      )}
    </Box>
  )
}

export default EducationAssistant
