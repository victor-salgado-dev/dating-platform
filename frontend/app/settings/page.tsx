'use client';

import { FormEvent, useEffect, useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';

import { apiFetch, ApiError } from '@/lib/api';
import { useI18n } from '@/lib/i18n/context';
import type { Dictionary } from '@/lib/i18n/dictionaries/es';
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
  const { dictionary } = useI18n();
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
        setDeleteError(dictionary.settings.errorWrongPassword);
      } else {
        setDeleteError(dictionary.settings.errorDeleteFailed);
      }
    } finally {
      setDeleting(false);
    }
  }

  if (loading) {
    return (
      <main className={styles.main}>
        <p>{dictionary.common.loading}</p>
      </main>
    );
  }

  if (!me) {
    return (
      <main className={styles.main}>
        <h1>{dictionary.settings.title}</h1>
        <p>
          <Link href="/login">{dictionary.settings.notLoggedInLogin}</Link>
          {dictionary.settings.notLoggedInBetween}
          <Link href="/register">{dictionary.settings.notLoggedInRegister}</Link>
          {dictionary.settings.notLoggedInEnd}
        </p>
      </main>
    );
  }

  return (
    <main className={styles.main}>
      <header className={styles.pageHeader}>
        <p className={styles.eyebrow}>{dictionary.settings.eyebrow}</p>
        <h1>{dictionary.settings.title}</h1>
        <p className={styles.intro}>{dictionary.settings.intro}</p>
      </header>

      <section className={styles.section}>
        <h2>{dictionary.settings.sectionAccount}</h2>
        <div className={styles.card}>
          <div className={styles.settingRow}>
            <div>
              <h3>{dictionary.common.fields.email}</h3>
              <p>{me.email}</p>
            </div>
            <span className={styles.status}>
              {me.email_verified ? dictionary.settings.statusVerified : dictionary.settings.statusPending}
            </span>
          </div>
          <div className={styles.settingRow}>
            <div>
              <h3>{dictionary.common.fields.password}</h3>
              <p>{dictionary.settings.passwordHint}</p>
            </div>
          </div>
        </div>
      </section>

      <section className={styles.section}>
        <h2>{dictionary.settings.sectionConsents}</h2>
        <div className={styles.card}>
          <Link href="/legal/privacy" className={styles.settingRow}>
            <div>
              <h3>{dictionary.settings.privacyPolicy}</h3>
              <p>{consentSummary(consents, 'privacy_policy', dictionary)}</p>
            </div>
            <span className={styles.chevron} aria-hidden="true">&gt;</span>
          </Link>
          <Link href="/legal/terms" className={styles.settingRow}>
            <div>
              <h3>{dictionary.settings.termsAndConditions}</h3>
              <p>{consentSummary(consents, 'terms', dictionary)}</p>
            </div>
            <span className={styles.chevron} aria-hidden="true">&gt;</span>
          </Link>
        </div>
      </section>

      <section className={styles.section}>
        <h2>{dictionary.settings.sectionPrivacy}</h2>
        <div className={styles.card}>
          <Link href="/legal/privacy" className={styles.settingRow}>
            <div>
              <h3>{dictionary.settings.privacyOptions}</h3>
              <p>{dictionary.settings.privacyOptionsHint}</p>
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
              <h3>{dictionary.settings.advanced}</h3>
              <p>{dictionary.settings.advancedHint}</p>
            </div>
            <span className={styles.chevron} aria-hidden="true">{advancedOpen ? '^' : '>'}</span>
          </button>
        </div>

        {advancedOpen && (
          <section className={styles.dangerZone}>
            <h4>{dictionary.settings.deleteAccount}</h4>
            <p>
              {dictionary.settings.deleteWarningPrefix}
              <Link href="/legal/privacy">{dictionary.legal.privacy.title}</Link>
              {dictionary.settings.deleteWarningSuffix}
            </p>

            {!confirmingDelete ? (
              <button type="button" onClick={() => setConfirmingDelete(true)} className={styles.deleteButton}>
                {dictionary.settings.deleteButton}
              </button>
            ) : (
              <form onSubmit={handleDelete} className={styles.deleteForm}>
                <label>
                  {dictionary.settings.confirmPassword}
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
                    {deleting ? dictionary.settings.deleting : dictionary.settings.confirmDeletion}
                  </button>
                  <button type="button" onClick={() => setConfirmingDelete(false)} disabled={deleting}>
                    {dictionary.settings.cancel}
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

function consentSummary(consents: ConsentRecord[] | null, documentType: string, dictionary: Dictionary) {
  const consent = consents?.find((item) => item.document_type === documentType);
  return consent
    ? dictionary.settings.consentAccepted.replace('{version}', consent.document_version)
    : dictionary.settings.consentCheck;
}
