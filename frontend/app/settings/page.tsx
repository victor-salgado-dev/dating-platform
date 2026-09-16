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

export default function SettingsPage() {
  const router = useRouter();
  const [me, setMe] = useState<Me | null>(null);
  const [consents, setConsents] = useState<ConsentRecord[] | null>(null);
  const [loading, setLoading] = useState(true);

  const [deletePassword, setDeletePassword] = useState('');
  const [deleteError, setDeleteError] = useState<string | null>(null);
  const [deleting, setDeleting] = useState(false);
  const [confirmingDelete, setConfirmingDelete] = useState(false);
  const [advancedOpen, setAdvancedOpen] = useState(false);

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

  async function handleDelete(event: FormEvent) {
    event.preventDefault();
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
        <h1>Ajustes</h1>
        <p>
          <Link href="/login">Inicia sesión</Link> o{' '}
          <Link href="/register">crea una cuenta</Link>.
        </p>
      </main>
    );
  }

  return (
    <main className={styles.main}>
      <header className={styles.pageHeader}>
        <p className={styles.eyebrow}>Cuenta</p>
        <h1>Ajustes</h1>
        <p className={styles.intro}>Gestiona la información y privacidad de tu cuenta.</p>
      </header>

      <section className={styles.section}>
        <h2>Cuenta</h2>
        <div className={styles.card}>
          <div className={styles.settingRow}>
            <div>
              <h3>Email</h3>
              <p>{me.email}</p>
            </div>
            <span className={styles.status}>{me.email_verified ? 'Verificado' : 'Pendiente'}</span>
          </div>
          <div className={styles.settingRow}>
            <div>
              <h3>Contraseña</h3>
              <p>Se solicita para confirmar acciones sensibles.</p>
            </div>
          </div>
        </div>
      </section>

      <section className={styles.section}>
        <h2>Consentimientos</h2>
        <div className={styles.card}>
          <Link href="/legal/privacy" className={styles.settingRow}>
            <div>
              <h3>Política de privacidad</h3>
              <p>{consentSummary(consents, 'privacy_policy')}</p>
            </div>
            <span className={styles.chevron} aria-hidden="true">&gt;</span>
          </Link>
          <Link href="/legal/terms" className={styles.settingRow}>
            <div>
              <h3>Términos y condiciones</h3>
              <p>{consentSummary(consents, 'terms')}</p>
            </div>
            <span className={styles.chevron} aria-hidden="true">&gt;</span>
          </Link>
        </div>
      </section>

      <section className={styles.section}>
        <h2>Privacidad</h2>
        <div className={styles.card}>
          <Link href="/legal/privacy" className={styles.settingRow}>
            <div>
              <h3>Opciones de privacidad</h3>
              <p>Consulta cómo se trata y protege tu información.</p>
            </div>
            <span className={styles.chevron} aria-hidden="true">&gt;</span>
          </Link>
          <button
            type="button"
            className={styles.settingRow}
            onClick={() => setAdvancedOpen((open) => !open)}
            aria-expanded={advancedOpen}
          >
            <div>
              <h3>Opciones avanzadas</h3>
              <p>Acciones permanentes de la cuenta.</p>
            </div>
            <span className={styles.chevron} aria-hidden="true">{advancedOpen ? '^' : '>'}</span>
          </button>
        </div>

        {advancedOpen && (
          <section className={styles.dangerZone}>
            <h4>Eliminar cuenta</h4>
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
                    onChange={(event) => setDeletePassword(event.target.value)}
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
        )}
      </section>
    </main>
  );
}

function consentSummary(consents: ConsentRecord[] | null, documentType: string) {
  const consent = consents?.find((item) => item.document_type === documentType);
  return consent ? `Aceptado · versión ${consent.document_version}` : 'Consulta el documento';
}