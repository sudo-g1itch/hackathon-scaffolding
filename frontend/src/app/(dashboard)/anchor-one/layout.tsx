'use client'

import React from 'react'

import { AnchorOneProvider } from '@/contexts/AnchorOneContext'

export default function AnchorOneLayout({ children }: { children: React.ReactNode }) {
  return <AnchorOneProvider>{children}</AnchorOneProvider>
}
