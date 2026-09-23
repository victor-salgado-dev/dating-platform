'use client';

import { FormEvent, useState } from 'react';

import { apiFetch, ApiError } from '@/lib/api';
import { useI18n } from '@/lib/i18n/context';
import styles from '../legal.module.css';
import formStyles from '../../login/page.module.css';

export default function ContactPage() {
  const { dictionary } = useI18n();
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [message, setMessage] = useState('');
  const [sent, setSent] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [sending, setSending] = useState(false);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setSending(true);
    setError(null);

    try {
      await apiFetch<void>('/contact', {
        method: 'POST',
        body: JSON.stringify({ name, email, message }),
      });
      setSent(true);
    } catch (err) {
      if (err instanceof ApiError && err.status === 429) {
        setError(dictionary.contact.errorRateLimited);
      } else {
        setError(dictionary.contact.errorGeneric);
      }
    } finally {
      setSending(false);
    }
  }

  return (
    <main className={styles.main}>
      <h1>{dictionary.contact.title}</h1>
      <p>{dictionary.contact.intro}</p>

      {sent ? (
        <p>{dictionary.contact.success}</p>
      ) : (
        <form onSubmit={handleSubmit} className={formStyles.form}>
          <label>
            {dictionary.contact.fieldName}
            <input type="text" value={name} onChange={(e) => setName(e.target.value)} required />
          </label>
          <label>
            {dictionary.common.fields.email}
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
            />
          </label>
          <label>
            {dictionary.contact.fieldMessage}
            <textarea
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              required
              rows={5}
              maxLength={3000}
            />
          </label>

          {error && <p className={formStyles.error}>{error}</p>}

          <button type="submit" disabled={sending}>
            {sending ? dictionary.contact.sending : dictionary.common.send}
          </button>
        </form>
      )}
    </main>
  );
}
