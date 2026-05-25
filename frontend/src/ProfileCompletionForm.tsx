import { useState } from 'react'
import { useAuth } from 'react-oidc-context'
import { backendClient, ApiError } from './api/client'

interface Props {
  onComplete: () => void
}

interface Fields {
  date_of_birth: string
  weight: string
  gender: string
}

interface Status {
  type: 'error'
  message: string
}

export default function ProfileCompletionForm({ onComplete }: Props) {
  const { user } = useAuth()
  const [fields, setFields] = useState<Fields>({ date_of_birth: '', weight: '', gender: '' })
  const [status, setStatus] = useState<Status | null>(null)
  const [loading, setLoading] = useState(false)

  function handleChange(e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) {
    setFields(f => ({ ...f, [e.target.name]: e.target.value }))
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setLoading(true)
    setStatus(null)

    const body = {
      auth_subject:  user?.profile.sub          ?? '',
      first_name:    user?.profile.given_name   ?? '',
      last_name:     user?.profile.family_name  ?? '',
      email:         user?.profile.email        ?? '',
      date_of_birth: fields.date_of_birth,
      weight:        Number(fields.weight),
      gender:        fields.gender,
    }

    try {
      await backendClient.createUser(body)
      onComplete()
    } catch (err) {
      if (err instanceof ApiError && err.status === 400) {
        setStatus({ type: 'error', message: err.message })
      } else {
        setStatus({ type: 'error', message: 'Something went wrong. Please try again.' })
      }
    } finally {
      setLoading(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} style={styles.form}>
      <h2>Complete your profile</h2>
      <p style={styles.hint}>A few more details before you get started.</p>

      <label style={styles.label}>
        Date of birth
        <input name="date_of_birth" type="date" value={fields.date_of_birth} onChange={handleChange} style={styles.input} />
      </label>

      <label style={styles.label}>
        Weight (kg)
        <input name="weight" type="number" step="0.01" min="0" value={fields.weight} onChange={handleChange} style={styles.input} />
      </label>

      <label style={styles.label}>
        Gender
        <select name="gender" value={fields.gender} onChange={handleChange} style={styles.input}>
          <option value="">Select…</option>
          <option value="male">Male</option>
          <option value="female">Female</option>
          <option value="unknown">Unknown</option>
        </select>
      </label>

      {status && (
        <p style={styles.error}>{status.message}</p>
      )}

      <button type="submit" disabled={loading} style={styles.button}>
        {loading ? 'Saving…' : 'Continue'}
      </button>
    </form>
  )
}

const styles: Record<string, React.CSSProperties> = {
  form:   { display: 'flex', flexDirection: 'column', gap: 12, maxWidth: 360 },
  hint:   { color: '#555', fontSize: 14, margin: 0 },
  label:  { display: 'flex', flexDirection: 'column', gap: 4, fontSize: 14 },
  input:  { padding: '6px 8px', fontSize: 14, borderRadius: 4, border: '1px solid #ccc' },
  button: { marginTop: 8, padding: '8px 16px', fontSize: 14, cursor: 'pointer' },
  error:  { color: '#9b2226', background: '#ffddd2', padding: '8px 12px', borderRadius: 4, margin: 0 },
}
