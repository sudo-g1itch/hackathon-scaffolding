'use client'

// My Caregiver — the user's side of the shared conversation.
//
// This is a human on the other end, not the AI coach. The distinction is worth
// stating on the page itself: the coach is private, this is not.
import { useRouter } from 'next/navigation'

import Alert from '@mui/material/Alert'
import Box from '@mui/material/Box'
import Button from '@mui/material/Button'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import CircularProgress from '@mui/material/CircularProgress'
import Typography from '@mui/material/Typography'

import SupportChat from '@components/anchor-one/SupportChat'
import { useAnchorOne } from '@/contexts/AnchorOneContext'
import { useAuth } from '@/contexts/AuthContext'

const MessagesPage = () => {
  const { user } = useAuth()
  const router = useRouter()
  const { dashboard, refreshDashboard } = useAnchorOne()

  const caregiverName = dashboard?.profile?.linked_caregiver_name

  if (!user) {
    return (
      <Box display='flex' justifyContent='center' alignItems='center' height='50vh'>
        <CircularProgress />
      </Box>
    )
  }

  return (
    <Box className='flex flex-col gap-6'>
      <Box>
        <Typography variant='h4' fontWeight={700}>
          My Caregiver
        </Typography>
        <Typography variant='body2' color='text.secondary'>
          {caregiverName
            ? `A direct line to ${caregiverName}.`
            : 'A direct line to the person supporting your recovery.'}
        </Typography>
      </Box>

      <Alert severity='info' icon={<i className='ri-information-line' />}>
        This conversation is with a real person. Your check-in recordings and your AI coach chats are never shown
        here — only what you choose to write.
      </Alert>

      <Card>
        <CardContent>
          <SupportChat
            patientId={user.id}
            onSent={() => void refreshDashboard()}
            unlinkedMessage='Choose a caregiver on your recovery plan, and you will be able to message them from here.'
            unlinkedAction={
              <Button
                variant='contained'
                onClick={() => router.push('/anchor-one/profile')}
                startIcon={<i className='ri-user-add-line' />}
              >
                Choose a caregiver
              </Button>
            }
          />
        </CardContent>
      </Card>
    </Box>
  )
}

export default MessagesPage
