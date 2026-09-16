'use client';

import { FormEvent, useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';

import { apiFetch, ApiError } from '@/lib/api';
import styles from '../login/page.module.css';

export default function RegisterPage() {
  const router = useRouter();
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
        setError('Ese email ya está registrado.');
      } else if (err instanceof ApiError && err.code === 'weak_password') {
        setError('La contraseña debe tener al menos 8 caracteres.');
      } else if (err instanceof ApiError && err.code === 'terms_not_accepted') {
        setError('Debes aceptar los Términos y la Política de Privacidad.');
      } else {
        setError('No se pudo completar el registro. Inténtalo de nuevo.');
      }
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className={styles.main}>
      <h1>Crear cuenta</h1>
      <p>Debes ser mayor de 18 años para registrarte.</p>

      <form onSubmit={handleSubmit} className={styles.form}>
        <label>
          Email
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            autoComplete="email"
          />
        </label>
        <label>
          Contraseña (mínimo 8 caracteres)
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
            He leído y acepto los <Link href="/legal/terms">Términos y Condiciones</Link> y la{' '}
            <Link href="/legal/privacy">Política de Privacidad</Link>.
          </span>
        </label>

        {error && <p className={styles.error}>{error}</p>}

        <button type="submit" disabled={loading || !acceptedTerms}>
          {loading ? 'Creando cuenta…' : 'Crear cuenta'}
        </button>
      </form>

      <p className={styles.switch}>
        ¿Ya tienes cuenta? <Link href="/login">Inicia sesión</Link>
      </p>
    </main>
  );
}
