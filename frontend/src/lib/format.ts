export function initials(firstName: string, lastName: string): string {
  const a = firstName.trim().charAt(0)
  const b = lastName.trim().charAt(0)
  return `${a}${b}`.toUpperCase() || '?'
}

/** A stable, vivid gradient per person, so faces in a list are told apart at a
 *  glance without needing photos. */
export function avatarGradient(seed: string): string {
  let hash = 0
  for (let i = 0; i < seed.length; i++) {
    hash = (hash * 31 + seed.charCodeAt(i)) | 0
  }
  const hue = Math.abs(hash) % 360
  // Second stop is offset around the wheel so every avatar has depth rather
  // than reading as a flat disc.
  return `linear-gradient(140deg, hsl(${hue} 72% 58%), hsl(${(hue + 48) % 360} 78% 46%))`
}
