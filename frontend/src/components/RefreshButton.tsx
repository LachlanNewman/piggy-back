'use client'

interface Props {
  busy: boolean
  onClick: () => void
}

export default function RefreshButton({ busy, onClick }: Props) {
  return (
    <button
      className={`btn btn-ghost btn-icon${busy ? ' is-busy' : ''}`}
      onClick={onClick}
      disabled={busy}
      aria-label={busy ? 'Refreshing carriers' : 'Refresh carriers'}
      title="Refresh"
    >
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" aria-hidden="true">
        <path d="M20 11A8 8 0 0 0 6.3 6.3L3.5 9" />
        <path d="M4 13a8 8 0 0 0 13.7 4.7l2.8-2.7" />
        <path d="M3.5 4v5h5M20.5 20v-5h-5" />
      </svg>
    </button>
  )
}
