'use client';

import { FormEvent, useEffect, useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';

import { apiFetch, ApiError } from '@/lib/api';
import styles from './page.module.css';

type Me = {
  id: string;
  email: string;
  email_verified: boolean;
  status: string;
  created_at: string;
};

type ConsentRecord = {
  document_type: string;
  document_version: string;
  accepted_at: string;
};

export default function AccountPage() {
  const router = useRouter();
  const [me, setMe] = useState<Me | null>(null);
  const [consents, setConsents] = useState<ConsentRecord[] | null>(null);
  const [loading, setLoading] = useState(true);
  const [loggedOut, setLoggedOut] = useState(false);

  const [deletePassword, setDeletePassword] = useState('');
  const [deleteError, setDeleteError] = useState<string | null>(null);
  const [deleting, setDeleting] = useState(false);
  const [confirmingDelete, setConfirmingDelete] = useState(false);

  useEffect(() => {
    apiFetch<Me>('/auth/me')
      .then((data) => {
        setMe(data);
        apiFetch<ConsentRecord[]>('/consents/me')
          .then(setConsents)
          .catch(() => setConsents([]));
      })
      .catch(() => setMe(null))
      .finally(() => setLoading(false));
  }, []);

  async function handleLogout() {
    try {
      await apiFetch<void>('/auth/logout', { method: 'POST' });
    } catch {
      // Si falla, igualmente reflejamos que ya no hay sesión localmente.
    }
    setMe(null);
    setLoggedOut(true);
    window.dispatchEvent(new Event('auth-change'));
  }

  async function handleDelete(e: FormEvent) {
    e.preventDefault();
    setDeleting(true);
    setDeleteError(null);

    try {
      await apiFetch<void>('/auth/account', {
        method: 'DELETE',
        body: JSON.stringify({ password: deletePassword }),
      });
      router.push('/');
    } catch (err) {
      if (err instanceof ApiError && err.status === 403) {
        setDeleteError('Contraseña incorrecta.');
      } else {
        setDeleteError('No se pudo eliminar la cuenta.');
      }
    } finally {
      setDeleting(false);
    }
  }

  if (loading) {
    return (
      <main className={styles.main}>
        <p>Cargando…</p>
      </main>
    );
  }

  if (!me) {
    return (
      <main className={styles.main}>
        <h1>Cuenta</h1>
        {loggedOut && <p>Has cerrado sesión.</p>}
        <p>
          <Link href="/login">Inicia sesión</Link> o{' '}
          <Link href="/register">crea una cuenta</Link>.
        </p>
      </main>
    );
  }

  return (
    <main className={styles.main}>
      <h1>Cuenta</h1>

      <dl className={styles.details}>
        <dt>Email</dt>
        <dd>{me.email}</dd>
        <dt>Email verificado</dt>
        <dd>{me.email_verified ? 'Sí' : 'No'}</dd>
        <dt>Miembro desde</dt>
        <dd>{new Date(me.created_at).toLocaleDateString()}</dd>
      </dl>

      <p>
        <Link href="/profile/edit" className={styles.profileLink}>Modificar perfil</Link>
      </p>

      <button type="button" onClick={handleLogout} className={styles.logoutButton}>
        Cerrar sesión
      </button>

      {consents && consents.length > 0 && (
        <section className={styles.section}>
          <h2>Tus consentimientos</h2>
          <ul className={styles.consentList}>
            {consents.map((c, i) => (
              <li key={i}>
                {c.document_type === 'terms' ? 'Términos y Condiciones' : 'Política de Privacidad'}{' '}
                (versión {c.document_version}) — aceptado el{' '}
                {new Date(c.accepted_at).toLocaleString()}
              </li>
            ))}
          </ul>
        </section>
      )}

      <section className={styles.dangerZone}>
        <h2>Eliminar cuenta</h2>
        <p>
          Esta acción es irreversible. Tu perfil, fotos y datos dejarán de estar disponibles.
          El histórico se conserva de forma limitada según se describe en la{' '}
          <Link href="/legal/privacy">Política de Privacidad</Link>.
        </p>

        {!confirmingDelete ? (
          <button type="button" onClick={() => setConfirmingDelete(true)} className={styles.deleteButton}>
            Eliminar mi cuenta
          </button>
        ) : (
          <form onSubmit={handleDelete} className={styles.deleteForm}>
            <label>
              Confirma tu contraseña
              <input
                type="password"
                value={deletePassword}
                onChange={(e) => setDeletePassword(e.target.value)}
                required
                autoComplete="current-password"
              />
            </label>
            {deleteError && <p className={styles.error}>{deleteError}</p>}
            <div className={styles.deleteActions}>
              <button type="submit" disabled={deleting} className={styles.deleteButton}>
                {deleting ? 'Eliminando…' : 'Confirmar eliminación'}
              </button>
              <button type="button" onClick={() => setConfirmingDelete(false)} disabled={deleting}>
                Cancelar
              </button>
            </div>
          </form>
        )}
      </section>
    </main>
  );
}
