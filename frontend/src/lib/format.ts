export function initials(firstName: string, lastName: string): string {
  const a = firstName.trim().charAt(0)
  const b = lastName.trim().charAt(0)
  return `${a}${b}`.toUpperCase() || '?'
}
