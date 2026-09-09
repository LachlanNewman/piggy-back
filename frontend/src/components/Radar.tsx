import PiggybackMark from './PiggybackMark'

/** Sonar sweep shown while we look for carriers — the app's one moment of
 *  spectacle, and an honest signal that it is actively scanning. */
export default function Radar() {
  return (
    <div className="radar" aria-hidden="true">
      <span className="radar-sweep" />
      <span className="radar-grid" />
      <span className="radar-ring" />
      <span className="radar-ring" />
      <span className="radar-ring" />
      <span className="radar-core"><PiggybackMark /></span>
    </div>
  )
}
