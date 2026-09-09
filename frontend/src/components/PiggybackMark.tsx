/** One figure carrying another — the app's mark, used at every size. */
export default function PiggybackMark({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 32 32" fill="none" aria-hidden="true">
      {/* carrier */}
      <circle cx="12" cy="8.5" r="3.6" fill="currentColor" />
      <path
        d="M12 13.5c-3 0-5 1.9-5.4 4.6L5.8 23a1.5 1.5 0 0 0 3 .5l.7-3.6V30a1.6 1.6 0 0 0 3.2 0v-6h.6v6a1.6 1.6 0 0 0 3.2 0V20l.7 3.6a1.5 1.5 0 0 0 3-.5l-.8-4.9c-.4-2.7-2.4-4.6-5.4-4.6Z"
        fill="currentColor"
      />
      {/* rider on the back */}
      <circle cx="22.5" cy="6" r="2.9" fill="currentColor" opacity=".62" />
      <path
        d="M22.5 10c-2.4 0-4.1 1.5-4.4 3.7l-.5 3.1a1.25 1.25 0 0 0 2.5.4l.4-2.3v3.4l3.9 3.9a1.4 1.4 0 0 0 2-2l-2.6-2.6v-2.7l.4 2.3a1.25 1.25 0 0 0 2.5-.4l-.5-3.1c-.3-2.2-2-3.7-4.4-3.7Z"
        fill="currentColor"
        opacity=".62"
      />
    </svg>
  )
}
