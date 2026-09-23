'use client';

import { FormEvent, useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';

import { apiFetch, ApiError } from '@/lib/api';
import { useI18n } from '@/lib/i18n/context';
import styles from './page.module.css';

export default function LoginPage() {
  const router = useRouter();
  const { dictionary } = useI18n();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      await apiFetch<{ id: string }>('/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
      });
      router.push('/discover');
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        setError(dictionary.auth.login.errorInvalid);
      } else if (err instanceof ApiError && err.status === 403) {
        setError(dictionary.auth.login.errorSuspended);
      } else {
        setError(dictionary.auth.login.errorGeneric);
      }
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className={styles.main}>
      <h1>{dictionary.auth.login.title}</h1>

      <form onSubmit={handleSubmit} className={styles.form}>
        <label>
          {dictionary.common.fields.email}
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            autoComplete="email"
          />
        </label>
        <label>
          {dictionary.common.fields.password}
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            autoComplete="current-password"
          />
        </label>

        {error && <p className={styles.error}>{error}</p>}

        <button type="submit" disabled={loading}>
          {loading ? dictionary.auth.login.loading : dictionary.auth.login.submit}
        </button>
      </form>

      <p className={styles.switch}>
        {dictionary.auth.login.noAccount}{' '}
        <Link href="/register">{dictionary.auth.register.title}</Link>
      </p>
    </main>
  );
}
