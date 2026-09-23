'use client';

import { useEffect, useState } from 'react';

import {
  apiFetch,
  ApiError,
  AdminUsersResponse,
  AdminReportsResponse,
} from '@/lib/api';
import { useI18n } from '@/lib/i18n/context';
import type { Dictionary } from '@/lib/i18n/dictionaries/es';
import type { Locale } from '@/lib/i18n/config';
import styles from './page.module.css';

export default function AdminPage() {
  const { dictionary } = useI18n();
  const [tab, setTab] = useState<'users' | 'reports'>('reports');
  const [forbidden, setForbidden] = useState(false);

  if (forbidden) {
    return (
      <main className={styles.main}>
        <h1>{dictionary.admin.title}</h1>
        <p className={styles.error}>{dictionary.admin.forbidden}</p>
      </main>
    );
  }

  return (
    <main className={styles.main}>
      <h1>{dictionary.admin.title}</h1>
      <div className={styles.tabs}>
        <button
          type="button"
          onClick={() => setTab('reports')}
          disabled={tab === 'reports'}
        >
          {dictionary.admin.tabReports}
        </button>
        <button type="button" onClick={() => setTab('users')} disabled={tab === 'users'}>
          {dictionary.admin.tabUsers}
        </button>
      </div>

      {tab === 'reports' ? (
        <ReportsPanel onForbidden={() => setForbidden(true)} />
      ) : (
        <UsersPanel onForbidden={() => setForbidden(true)} />
      )}
    </main>
  );
}

function ReportsPanel({ onForbidden }: { onForbidden: () => void }) {
  const { locale, dictionary } = useI18n();
  const [data, setData] = useState<AdminReportsResponse | null>(null);
  const [statusFilter, setStatusFilter] = useState('pending');
  const [error, setError] = useState<string | null>(null);
  const [busyId, setBusyId] = useState<string | null>(null);

  function load() {
    setError(null);
    apiFetch<AdminReportsResponse>(`/admin/reports?status=${statusFilter}`)
      .then(setData)
      .catch((err: unknown) => {
        if (err instanceof ApiError && err.status === 403) {
          onForbidden();
        } else {
          setError(dictionary.admin.reportsLoadError);
        }
      });
  }

  // eslint-disable-next-line react-hooks/exhaustive-deps
  useEffect(load, [statusFilter]);

  async function resolve(id: string, status: 'reviewed' | 'dismissed') {
    setBusyId(id);
    try {
      await apiFetch<void>(`/admin/reports/${id}/resolve`, {
        method: 'POST',
        body: JSON.stringify({ status }),
      });
      load();
    } catch {
      setError(dictionary.admin.reportsResolveError);
    } finally {
      setBusyId(null);
    }
  }

  return (
    <section>
      <label className={styles.filter}>
        {dictionary.admin.statusLabel}{' '}
        <select value={statusFilter} onChange={(e) => setStatusFilter(e.target.value)}>
          <option value="pending">{dictionary.admin.statusPending}</option>
          <option value="reviewed">{dictionary.admin.statusReviewed}</option>
          <option value="dismissed">{dictionary.admin.statusDismissed}</option>
        </select>
      </label>

      {error && <p className={styles.error}>{error}</p>}

      <table className={styles.table}>
        <thead>
          <tr>
            <th>{dictionary.admin.tableReason}</th>
            <th>{dictionary.admin.tableReported}</th>
            <th>{dictionary.admin.tableReportedBy}</th>
            <th>{dictionary.admin.tableDetails}</th>
            <th>{dictionary.admin.tableDate}</th>
            {statusFilter === 'pending' && <th>{dictionary.admin.tableActions}</th>}
          </tr>
        </thead>
        <tbody>
          {data?.items.map((r) => (
            <tr key={r.id}>
              <td>{r.reason}</td>
              <td>{r.reported_name ?? r.reported_id}</td>
              <td>{r.reporter_name ?? r.reporter_id}</td>
              <td>{r.description ?? '—'}</td>
              <td>{formatDate(r.created_at, locale)}</td>
              {statusFilter === 'pending' && (
                <td className={styles.actions}>
                  <button
                    type="button"
                    disabled={busyId === r.id}
                    onClick={() => resolve(r.id, 'reviewed')}
                  >
                    {dictionary.admin.actionMarkReviewed}
                  </button>
                  <button
                    type="button"
                    disabled={busyId === r.id}
                    onClick={() => resolve(r.id, 'dismissed')}
                  >
                    {dictionary.admin.actionDismiss}
                  </button>
                </td>
              )}
            </tr>
          ))}
        </tbody>
      </table>
      {data && data.items.length === 0 && <p>{dictionary.admin.noReports}</p>}
    </section>
  );
}

function UsersPanel({ onForbidden }: { onForbidden: () => void }) {
  const { locale, dictionary } = useI18n();
  const [data, setData] = useState<AdminUsersResponse | null>(null);
  const [statusFilter, setStatusFilter] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [busyId, setBusyId] = useState<string | null>(null);

  function load() {
    setError(null);
    const qs = statusFilter ? `?status=${statusFilter}` : '';
    apiFetch<AdminUsersResponse>(`/admin/users${qs}`)
      .then(setData)
      .catch((err: unknown) => {
        if (err instanceof ApiError && err.status === 403) {
          onForbidden();
        } else {
          setError(dictionary.admin.usersLoadError);
        }
      });
  }

  // eslint-disable-next-line react-hooks/exhaustive-deps
  useEffect(load, [statusFilter]);

  async function setSuspended(id: string, suspend: boolean) {
    setBusyId(id);
    try {
      await apiFetch<void>(`/admin/users/${id}/${suspend ? 'suspend' : 'reactivate'}`, {
        method: 'POST',
      });
      load();
    } catch {
      setError(dictionary.admin.usersUpdateError);
    } finally {
      setBusyId(null);
    }
  }

  return (
    <section>
      <label className={styles.filter}>
        {dictionary.admin.statusLabel}{' '}
        <select value={statusFilter} onChange={(e) => setStatusFilter(e.target.value)}>
          <option value="">{dictionary.admin.statusAll}</option>
          <option value="active">{dictionary.admin.statusActive}</option>
          <option value="suspended">{dictionary.admin.statusSuspended}</option>
          <option value="deleted">{dictionary.admin.statusDeleted}</option>
        </select>
      </label>

      {error && <p className={styles.error}>{error}</p>}

      <table className={styles.table}>
        <thead>
          <tr>
            <th>{dictionary.admin.tableEmail}</th>
            <th>{dictionary.admin.tableStatus}</th>
            <th>{dictionary.admin.tableRole}</th>
            <th>{dictionary.admin.tableVerified}</th>
            <th>{dictionary.admin.tableCreated}</th>
            <th>{dictionary.admin.tableActions}</th>
          </tr>
        </thead>
        <tbody>
          {data?.items.map((u) => (
            <tr key={u.id}>
              <td>{u.email}</td>
              <td>{u.status}</td>
              <td>{u.role}</td>
              <td>{u.email_verified ? dictionary.common.yes : dictionary.common.no}</td>
              <td>{formatDate(u.created_at, locale)}</td>
              <td className={styles.actions}>
                {u.status === 'active' && (
                  <button
                    type="button"
                    disabled={busyId === u.id}
                    onClick={() => setSuspended(u.id, true)}
                  >
                    {dictionary.admin.actionSuspend}
                  </button>
                )}
                {u.status === 'suspended' && (
                  <button
                    type="button"
                    disabled={busyId === u.id}
                    onClick={() => setSuspended(u.id, false)}
                  >
                    {dictionary.admin.actionReactivate}
                  </button>
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  );
}

function formatDate(value: string, locale: Locale, dictionary?: Dictionary) {
  void dictionary; // firma homogénea por si más adelante queremos formatear "hoy/ayer"
  return new Date(value).toLocaleDateString(locale === 'en' ? 'en-GB' : 'es-ES');
}
