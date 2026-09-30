'use client';

import { useEffect, useState, FormEvent } from 'react';
import Link from 'next/link';
import { useParams, useRouter } from 'next/navigation';

import { apiFetch, ApiError, MessagesResponse, MessageItem } from '@/lib/api';
import { useI18n } from '@/lib/i18n/context';
import styles from './page.module.css';

function formatMessageTime(iso: string): string {
  const date = new Date(iso);
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

export default function ConversationPage() {
  const params = useParams<{ id: string }>();
  const router = useRouter();
  const { dictionary } = useI18n();
  const [messages, setMessages] = useState<MessageItem[] | null>(null);
  const [draft, setDraft] = useState('');
  const [sending, setSending] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!params?.id) return;

    setLoading(true);
    setError(null);

    // Cargar la conversación también la marca como leída en el backend
    // (efecto secundario de GET .../messages, ver Fase 8 en el README).
    apiFetch<MessagesResponse>(`/messages/conversations/${params.id}/messages?page_size=50`)
      .then((res) => setMessages(res.items))
      .catch((err: unknown) => {
        if (err instanceof ApiError && err.status === 404) {
          setError(dictionary.conversation.notFound);
        } else if (err instanceof ApiError && err.status === 401) {
          const returnTo = window.location.pathname + window.location.search;
          router.replace(`/login?next=${encodeURIComponent(returnTo)}`);
        } else {
          setError(dictionary.conversation.loadError);
        }
      })
      .finally(() => setLoading(false));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [params?.id]);

  async function handleSend(e: FormEvent) {
    e.preventDefault();
    if (!params?.id || !draft.trim() || sending) return;

    setSending(true);
    try {
      const sent = await apiFetch<MessageItem>(`/messages/conversations/${params.id}/messages`, {
        method: 'POST',
        body: JSON.stringify({ body: draft }),
      });
      setMessages((prev) => [...(prev ?? []), sent]);
      setDraft('');
    } catch {
      setError(dictionary.conversation.sendError);
    } finally {
      setSending(false);
    }
  }

  return (
    <main className={styles.main}>
      <Link href="/messages" className={styles.back}>
        {dictionary.conversation.back}
      </Link>

      {loading && <p>{dictionary.common.loading}</p>}
      {error && <p className={styles.error}>{error}</p>}

      {messages && (
        <>
          <ul className={styles.thread}>
            {messages.map((m) => (
              <li
                key={m.id}
                className={m.is_mine ? styles.bubbleMine : styles.bubbleTheirs}
              >
                <span className={styles.body}>{m.body}</span>
                <span className={styles.meta}>
                  <span className={styles.time}>{formatMessageTime(m.created_at)}</span>
                  {m.is_mine && (
                    <span className={`${styles.checks}${m.read_at ? ` ${styles.readChecks}` : ''}`}>
                      {m.read_at ? '✓✓' : '✓'}
                    </span>
                  )}
                </span>
              </li>
            ))}
          </ul>

          <form onSubmit={handleSend} className={styles.composer}>
            <input
              type="text"
              value={draft}
              onChange={(e) => setDraft(e.target.value)}
              placeholder={dictionary.conversation.placeholder}
              maxLength={2000}
              disabled={sending}
            />
            <button type="submit" disabled={sending || !draft.trim()}>
              {dictionary.common.send}
            </button>
          </form>
        </>
      )}
    </main>
  );
}
