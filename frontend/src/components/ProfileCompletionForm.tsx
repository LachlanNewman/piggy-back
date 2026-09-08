'use client'

import { useState } from 'react'
import { useAuth } from 'react-oidc-context'
import { backendClient, ApiError } from '@/lib/api/client'

interface Props {
  onComplete: () => void
}

interface Fields {
  date_of_birth: string
  weight: string
  gender: string
}

export default function ProfileCompletionForm({ onComplete }: Props) {
  const { user } = useAuth()
  const [fields, setFields] = useState<Fields>({ date_of_birth: '', weight: '', gender: '' })
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  function handleChange(e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) {
    setFields(f => ({ ...f, [e.target.name]: e.target.value }))
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setLoading(true)
    setError(null)

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
        setError(err.message)
      } else {
        setError('Something went wrong. Please try again.')
      }
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="card">
      <h2 className="card-title">Finish your profile</h2>
      <p className="muted">
        Your weight is how carriers decide whether they can take you. Be honest — backs are at stake.
      </p>

      <form onSubmit={handleSubmit} className="form" style={{ marginTop: 18 }}>
        <label className="field">
          <span>Date of birth</span>
          <input
            className="input"
            name="date_of_birth"
            type="date"
            value={fields.date_of_birth}
            onChange={handleChange}
          />
        </label>

        <label className="field">
          <span>Weight (kg)</span>
          <input
            className="input"
            name="weight"
            type="number"
            step="0.01"
            min="0"
            placeholder="72"
            value={fields.weight}
            onChange={handleChange}
          />
        </label>

        <label className="field">
          <span>Gender</span>
          <select className="input" name="gender" value={fields.gender} onChange={handleChange}>
            <option value="">Select…</option>
            <option value="male">Male</option>
            <option value="female">Female</option>
            <option value="unknown">Unknown</option>
          </select>
        </label>

        {error && <p className="banner banner-error">{error}</p>}

        <button type="submit" className="btn btn-block" disabled={loading}>
          {loading ? 'Saving…' : 'Continue'}
        </button>
      </form>
    </div>
  )
}
