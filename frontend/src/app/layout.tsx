import type { Metadata, Viewport } from 'next'
import Providers from './providers'
import './globals.css'

export const metadata: Metadata = {
  title: 'Piggy Back — piggyback rides on demand',
  description: 'Request a piggyback ride from someone nearby. Like a rideshare, but the vehicle is a person.',
}

export const viewport: Viewport = {
  width: 'device-width',
  initialScale: 1,
  themeColor: '#12100e',
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <Providers>{children}</Providers>
      </body>
    </html>
  )
}
