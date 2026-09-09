'use client'

import { initials, avatarGradient } from '@/lib/format'

const R = 26
const CIRCUMFERENCE = 2 * Math.PI * R

interface Props {
  firstName: string
  lastName: string
  /** 0–1 of the request's life still remaining. Omit for a plain avatar. */
  remaining?: number
}

export default function Avatar({ firstName, lastName, remaining }: Props) {
  const face = (
    <div
      className="avatar"
      style={{ backgroundImage: avatarGradient(`${firstName} ${lastName}`) }}
      aria-hidden="true"
    >
      {initials(firstName, lastName)}
    </div>
  )

  if (remaining === undefined) return face

  const clamped = Math.min(1, Math.max(0, remaining))
  return (
    <div className="avatar-timer">
      <svg className="ring" viewBox="0 0 56 56" aria-hidden="true">
        <circle className="ring-track" cx="28" cy="28" r={R} />
        <circle
          className={`ring-progress${clamped < 0.25 ? ' is-urgent' : ''}`}
          cx="28"
          cy="28"
          r={R}
          strokeDasharray={CIRCUMFERENCE}
          strokeDashoffset={CIRCUMFERENCE * (1 - clamped)}
        />
      </svg>
      {face}
    </div>
  )
}
