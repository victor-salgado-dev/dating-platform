'use client';

import { useEffect, useState } from 'react';

import {
  apiFetch,
  ApiError,
  AdminUsersResponse,
  AdminReportsResponse,
} from '@/lib/api';
import styles from './page.module.css';

export default function AdminPage() {
  const [tab, setTab] = useState<'users' | 'reports'>('reports');
  const [forbidden, setForbidden] = useState(false);

  if (forbidden) {
    return (
      <main className={styles.main}>
        <h1>Administración</h1>
        <p className={styles.error}>No tienes permisos de administración.</p>
      </main>
    );
  }

  return (
    <main className={styles.main}>
      <h1>Administración</h1>
      <div className={styles.tabs}>
        <button
          type="button"
          onClick={() => setTab('reports')}
          disabled={tab === 'reports'}
        >
          Reportes
        </button>
        <button type="button" onClick={() => setTab('users')} disabled={tab === 'users'}>
          Usuarios
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
          setError('No se pudieron cargar los reportes.');
        }
      });
  }

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
      setError('No se pudo resolver el reporte.');
    } finally {
      setBusyId(null);
    }
  }

  return (
    <section>
      <label className={styles.filter}>
        Estado:{' '}
        <select value={statusFilter} onChange={(e) => setStatusFilter(e.target.value)}>
          <option value="pending">Pendientes</option>
          <option value="reviewed">Revisados</option>
          <option value="dismissed">Descartados</option>
        </select>
      </label>

      {error && <p className={styles.error}>{error}</p>}

      <table className={styles.table}>
        <thead>
          <tr>
            <th>Motivo</th>
            <th>Reportado</th>
            <th>Reportado por</th>
            <th>Detalles</th>
            <th>Fecha</th>
            {statusFilter === 'pending' && <th>Acciones</th>}
          </tr>
        </thead>
        <tbody>
          {data?.items.map((r) => (
            <tr key={r.id}>
              <td>{r.reason}</td>
              <td>{r.reported_name ?? r.reported_id}</td>
              <td>{r.reporter_name ?? r.reporter_id}</td>
              <td>{r.description ?? '—'}</td>
              <td>{new Date(r.created_at).toLocaleDateString()}</td>
              {statusFilter === 'pending' && (
                <td className={styles.actions}>
                  <button
                    type="button"
                    disabled={busyId === r.id}
                    onClick={() => resolve(r.id, 'reviewed')}
                  >
                    Marcar revisado
                  </button>
                  <button
                    type="button"
                    disabled={busyId === r.id}
                    onClick={() => resolve(r.id, 'dismissed')}
                  >
                    Descartar
                  </button>
                </td>
              )}
            </tr>
          ))}
        </tbody>
      </table>
      {data && data.items.length === 0 && <p>No hay reportes en este estado.</p>}
    </section>
  );
}

function UsersPanel({ onForbidden }: { onForbidden: () => void }) {
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
          setError('No se pudieron cargar los usuarios.');
        }
      });
  }

  useEffect(load, [statusFilter]);

  async function setSuspended(id: string, suspend: boolean) {
    setBusyId(id);
    try {
      await apiFetch<void>(`/admin/users/${id}/${suspend ? 'suspend' : 'reactivate'}`, {
        method: 'POST',
      });
      load();
    } catch {
      setError('No se pudo actualizar el usuario.');
    } finally {
      setBusyId(null);
    }
  }

  return (
    <section>
      <label className={styles.filter}>
        Estado:{' '}
        <select value={statusFilter} onChange={(e) => setStatusFilter(e.target.value)}>
          <option value="">Todos</option>
          <option value="active">Activos</option>
          <option value="suspended">Suspendidos</option>
          <option value="deleted">Eliminados</option>
        </select>
      </label>

      {error && <p className={styles.error}>{error}</p>}

      <table className={styles.table}>
        <thead>
          <tr>
            <th>Email</th>
            <th>Estado</th>
            <th>Rol</th>
            <th>Verificado</th>
            <th>Alta</th>
            <th>Acciones</th>
          </tr>
        </thead>
        <tbody>
          {data?.items.map((u) => (
            <tr key={u.id}>
              <td>{u.email}</td>
              <td>{u.status}</td>
              <td>{u.role}</td>
              <td>{u.email_verified ? 'Sí' : 'No'}</td>
              <td>{new Date(u.created_at).toLocaleDateString()}</td>
              <td className={styles.actions}>
                {u.status === 'active' && (
                  <button type="button" disabled={busyId === u.id} onClick={() => setSuspended(u.id, true)}>
                    Suspender
                  </button>
                )}
                {u.status === 'suspended' && (
                  <button type="button" disabled={busyId === u.id} onClick={() => setSuspended(u.id, false)}>
                    Reactivar
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
