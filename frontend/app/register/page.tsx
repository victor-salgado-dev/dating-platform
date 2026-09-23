'use client';

import { FormEvent, useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';

import { apiFetch, ApiError } from '@/lib/api';
import { useI18n } from '@/lib/i18n/context';
import styles from '../login/page.module.css';

export default function RegisterPage() {
  const router = useRouter();
  const { dictionary } = useI18n();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [acceptedTerms, setAcceptedTerms] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      await apiFetch<{ id: string }>('/auth/register', {
        method: 'POST',
        body: JSON.stringify({ email, password, accepted_terms: acceptedTerms }),
      });
      router.push('/profile/edit');
    } catch (err) {
      if (err instanceof ApiError && err.code === 'email_taken') {
        setError(dictionary.auth.register.errorEmailTaken);
      } else if (err instanceof ApiError && err.code === 'weak_password') {
        setError(dictionary.auth.register.errorWeakPassword);
      } else if (err instanceof ApiError && err.code === 'terms_not_accepted') {
        setError(dictionary.auth.register.errorTermsNotAccepted);
      } else {
        setError(dictionary.auth.register.errorGeneric);
      }
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className={styles.main}>
      <h1>{dictionary.auth.register.title}</h1>
      <p>{dictionary.auth.register.ageNotice}</p>

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
          {dictionary.auth.register.passwordLabel}
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            minLength={8}
            autoComplete="new-password"
          />
        </label>

        <label className={styles.checkboxRow}>
          <input
            type="checkbox"
            checked={acceptedTerms}
            onChange={(e) => setAcceptedTerms(e.target.checked)}
            required
          />
          <span>
            {dictionary.auth.register.termsPrefix}{' '}
            <Link href="/legal/terms">{dictionary.legal.terms.title}</Link>{' '}
            {dictionary.auth.register.termsAnd}{' '}
            <Link href="/legal/privacy">{dictionary.legal.privacy.title}</Link>.
          </span>
        </label>

        {error && <p className={styles.error}>{error}</p>}

        <button type="submit" disabled={loading || !acceptedTerms}>
          {loading ? dictionary.auth.register.loading : dictionary.auth.register.submit}
        </button>
      </form>

      <p className={styles.switch}>
        {dictionary.auth.register.haveAccount}{' '}
        <Link href="/login">{dictionary.auth.login.link}</Link>
      </p>
    </main>
  );
}
