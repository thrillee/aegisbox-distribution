import React from "react"
import type { Metadata } from 'next'
import { Geist, Geist_Mono } from 'next/font/google'
import { ClerkProvider } from '@clerk/nextjs'
import './globals.css'

const _geist = Geist({ subsets: ["latin"] });
const _geistMono = Geist_Mono({ subsets: ["latin"] });

export const metadata: Metadata = {
  title: 'Aegisbox SMS Gateway',
  description: 'SMS management platform for service providers and administrators',
  generator: 'v0.app',
  icons: {
    icon: [
      {
        url: '/icon-light-32x32.png',
        media: '(prefers-color-scheme: light)',
      },
      {
        url: '/icon-dark-32x32.png',
        media: '(prefers-color-scheme: dark)',
      },
      {
        url: '/icon.svg',
        type: 'image/svg+xml',
      },
    ],
    apple: '/apple-icon.png',
  },
}

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <ClerkProvider
      appearance={{
        variables: {
          colorPrimary: '#ffffff',
          colorBackground: '#18181b',
          colorInputBackground: '#27272a',
          colorInputText: '#ffffff',
          colorText: '#ffffff',
          colorTextSecondary: '#a1a1aa',
          colorDanger: '#ef4444',
          colorSuccess: '#22c55e',
          colorWarning: '#f59e0b',
          colorBorder: '#3f3f46',
          colorNeutral: '#71717a',
          colorPrimaryForeground: '#000000',
        },
        elements: {
          card: 'bg-zinc-900 border border-white/10 shadow-xl',
          headerTitle: 'text-white text-xl font-semibold',
          headerSubtitle: 'text-zinc-400',
          formFieldLabel: 'text-zinc-300',
          formFieldInput: 'bg-zinc-800 border-zinc-700 text-white focus:border-white focus:ring-white/20',
          footerActionLink: 'text-white hover:text-zinc-300',
          identityPreviewText: 'text-white',
          identityPreviewArrow: 'text-zinc-400',
          dividerLine: 'bg-zinc-700',
          dividerText: 'text-zinc-400',
          socialButtonsBlockButton: 'bg-zinc-800 border-zinc-700 text-white hover:bg-zinc-700',
          socialButtonsBlockButtonArrow: 'text-zinc-400',
          formButtonPrimary: 'bg-white text-black hover:bg-zinc-200',
          footer: 'border-t border-zinc-800',
        },
      }}
    >
      <html lang="en">
        <body className={`${_geist.className} ${_geistMono.className} font-sans antialiased bg-black text-white`}>
          {children}
        </body>
      </html>
    </ClerkProvider>
  )
}
