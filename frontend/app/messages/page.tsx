'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';

import { apiFetch, ApiError, ConversationsResponse } from '@/lib/api';
import { useI18n } from '@/lib/i18n/context';
import styles from './page.module.css';

export default function MessagesPage() {
  const { dictionary } = useI18n();
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
          setError(dictionary.messages.errorUnauthorized);
        } else {
          setError(dictionary.messages.loadError);
        }
      })
      .finally(() => setLoading(false));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <main className={styles.main}>
      <h1>{dictionary.messages.title}</h1>

      {loading && <p>{dictionary.common.loading}</p>}
      {error && <p className={styles.error}>{error}</p>}

      {data && (
        <>
          {data.items.length === 0 ? (
            <p>{dictionary.messages.empty}</p>
          ) : (
            <ul className={styles.list}>
              {data.items.map((c) => (
                <li key={c.conversation_id}>
                  <Link href={`/messages/${c.conversation_id}`} className={styles.row}>
                    <div className={styles.rowMain}>
                      <span className={styles.name}>{c.other_participant.display_name}</span>
                      <span className={styles.preview}>
                        {c.last_message.is_mine ? dictionary.messages.youPrefix : ''}
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
