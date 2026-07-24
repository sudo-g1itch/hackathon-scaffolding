'use client'

// Generic confirm dialog used across list/detail screens.
import Button from '@mui/material/Button'
import Dialog from '@mui/material/Dialog'
import DialogActions from '@mui/material/DialogActions'
import DialogContent from '@mui/material/DialogContent'
import Typography from '@mui/material/Typography'

type ConfirmDialogProps = {
  open: boolean
  title: string
  message?: string
  confirmText?: string
  cancelText?: string
  loading?: boolean
  onConfirm: () => void
  onClose: () => void
}

const ConfirmDialog = ({
  open,
  title,
  message,
  confirmText = 'Confirm',
  cancelText = 'Cancel',
  loading = false,
  onConfirm,
  onClose
}: ConfirmDialogProps) => (
  <Dialog fullWidth maxWidth='xs' open={open} onClose={onClose} closeAfterTransition={false}>
    <DialogContent className='flex items-center flex-col text-center sm:pbs-10 sm:pbe-6 sm:pli-10'>
      <i className='ri-error-warning-line text-[64px] mbe-4 text-warning' />
      <Typography variant='h5'>{title}</Typography>
      {message ? (
        <Typography color='text.secondary' className='mbs-2'>
          {message}
        </Typography>
      ) : null}
    </DialogContent>
    <DialogActions className='justify-center pbs-0 sm:pbe-10 sm:pli-10 gap-2'>
      <Button variant='outlined' color='secondary' onClick={onClose} disabled={loading}>
        {cancelText}
      </Button>
      <Button variant='contained' color='error' onClick={onConfirm} disabled={loading}>
        {confirmText}
      </Button>
    </DialogActions>
  </Dialog>
)

export default ConfirmDialog
