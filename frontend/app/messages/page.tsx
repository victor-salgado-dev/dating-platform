'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';

import { apiFetch, ApiError, ConversationsResponse } from '@/lib/api';
import styles from './page.module.css';

export default function MessagesPage() {
  const [data, setData] = useState<ConversationsResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setLoading(true);
    setError(null);

    apiFetch<ConversationsResponse>('/messages/conversations')
      .then(setData)
      .catch((err: unknown) => {
        if (err instanceof ApiError && err.status === 401) {
          setError('Inicia sesión para ver tus mensajes.');
        } else {
          setError('No se pudieron cargar las conversaciones.');
        }
      })
      .finally(() => setLoading(false));
  }, []);

  return (
    <main className={styles.main}>
      <h1>Mensajes</h1>

      {loading && <p>Cargando…</p>}
      {error && <p className={styles.error}>{error}</p>}

      {data && (
        <>
          {data.items.length === 0 ? (
            <p>Todavía no tienes conversaciones.</p>
          ) : (
            <ul className={styles.list}>
              {data.items.map((c) => (
                <li key={c.conversation_id}>
                  <Link href={`/messages/${c.conversation_id}`} className={styles.row}>
                    <div className={styles.rowMain}>
                      <span className={styles.name}>{c.other_participant.display_name}</span>
                      <span className={styles.preview}>
                        {c.last_message.is_mine ? 'Tú: ' : ''}
                        {c.last_message.body}
                      </span>
                    </div>
                    {c.unread_count > 0 && (
                      <span className={styles.badge}>{c.unread_count}</span>
                    )}
                  </Link>
                </li>
              ))}
            </ul>
          )}
        </>
      )}
    </main>
  );
}
