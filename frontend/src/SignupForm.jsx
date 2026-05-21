import { useState } from 'react'

const EMPTY = { first_name: '', last_name: '', email: '', date_of_birth: '', weight: '', gender: '' }

export default function SignupForm() {
  const [fields, setFields] = useState(EMPTY)
  const [status, setStatus] = useState(null) // { type: 'success' | 'error', message: string }
  const [loading, setLoading] = useState(false)

  function handleChange(e) {
    setFields(f => ({ ...f, [e.target.name]: e.target.value }))
  }

  async function handleSubmit(e) {
    e.preventDefault()
    setLoading(true)
    setStatus(null)

    const body = { ...fields, weight: Number(fields.weight) }

    try {
      const res = await fetch('/api/v1/users', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      })

      const data = await res.json()

      if (res.status === 201) {
        setStatus({ type: 'success', message: 'Account created!' })
        setFields(EMPTY)
      } else if (res.status === 400 || res.status === 409) {
        setStatus({ type: 'error', message: data.error })
      } else {
        setStatus({ type: 'error', message: 'Something went wrong. Please try again.' })
      }
    } catch {
      setStatus({ type: 'error', message: 'Something went wrong. Please try again.' })
    } finally {
      setLoading(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} style={styles.form}>
      <h2>Create account</h2>

      <label style={styles.label}>
        First name
        <input name="first_name" value={fields.first_name} onChange={handleChange} style={styles.input} />
      </label>

      <label style={styles.label}>
        Last name
        <input name="last_name" value={fields.last_name} onChange={handleChange} style={styles.input} />
      </label>

      <label style={styles.label}>
        Email
        <input name="email" type="email" value={fields.email} onChange={handleChange} style={styles.input} />
      </label>

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
        <p style={status.type === 'success' ? styles.success : styles.error}>
          {status.message}
        </p>
      )}

      <button type="submit" disabled={loading} style={styles.button}>
        {loading ? 'Submitting…' : 'Sign up'}
      </button>
    </form>
  )
}

const styles = {
  form: { display: 'flex', flexDirection: 'column', gap: 12, maxWidth: 360 },
  label: { display: 'flex', flexDirection: 'column', gap: 4, fontSize: 14 },
  input: { padding: '6px 8px', fontSize: 14, borderRadius: 4, border: '1px solid #ccc' },
  button: { marginTop: 8, padding: '8px 16px', fontSize: 14, cursor: 'pointer' },
  success: { color: '#2d6a4f', background: '#d8f3dc', padding: '8px 12px', borderRadius: 4, margin: 0 },
  error: { color: '#9b2226', background: '#ffddd2', padding: '8px 12px', borderRadius: 4, margin: 0 },
}
