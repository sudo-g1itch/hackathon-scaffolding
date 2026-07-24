'use client'

import { useEffect, useRef, useState } from 'react'

import type { TextFieldProps } from '@mui/material/TextField'
import TextField from '@mui/material/TextField'

type DebouncedInputProps = Omit<TextFieldProps, 'onChange'> & {
  value: string | number
  onChange: (value: string | number) => void
  debounce?: number
}

const DebouncedInput = ({
  value: initialValue,
  onChange,
  debounce = 500,
  ...props
}: DebouncedInputProps) => {
  const [value, setValue] = useState(initialValue)

  const onChangeRef = useRef(onChange)

  useEffect(() => {
    onChangeRef.current = onChange
  })

  useEffect(() => {
    setValue(initialValue)
  }, [initialValue])

  useEffect(() => {
    if (value === initialValue) return

    const timeout = setTimeout(() => {
      onChangeRef.current(value)
    }, debounce)

    return () => clearTimeout(timeout)
  }, [value, initialValue, debounce])

  return (
    <TextField
      {...props}
      value={value}
      onChange={e => setValue(e.target.value)}
      inputProps={{ ...props.inputProps, suppressHydrationWarning: true }}
      InputLabelProps={{ ...props.InputLabelProps, suppressHydrationWarning: true }}
    />
  )
}

export default DebouncedInput
