'use client';

import { useEffect, useState, FormEvent } from 'react';
import Link from 'next/link';
import { useParams } from 'next/navigation';

import { apiFetch, ApiError, MessagesResponse, MessageItem } from '@/lib/api';
import styles from './page.module.css';

export default function ConversationPage() {
  const params = useParams<{ id: string }>();
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
          setError('Esta conversación no existe.');
        } else if (err instanceof ApiError && err.status === 401) {
          setError('Inicia sesión para ver esta conversación.');
        } else {
          setError('No se pudieron cargar los mensajes.');
        }
      })
      .finally(() => setLoading(false));
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
      setError('No se pudo enviar el mensaje.');
    } finally {
      setSending(false);
    }
  }

  return (
    <main className={styles.main}>
      <Link href="/messages" className={styles.back}>
        &larr; Todas las conversaciones
      </Link>

      {loading && <p>Cargando…</p>}
      {error && <p className={styles.error}>{error}</p>}

      {messages && (
        <>
          <ul className={styles.thread}>
            {messages.map((m) => (
              <li
                key={m.id}
                className={m.is_mine ? styles.bubbleMine : styles.bubbleTheirs}
              >
                {m.body}
              </li>
            ))}
          </ul>

          <form onSubmit={handleSend} className={styles.composer}>
            <input
              type="text"
              value={draft}
              onChange={(e) => setDraft(e.target.value)}
              placeholder="Escribe un mensaje…"
              maxLength={2000}
              disabled={sending}
            />
            <button type="submit" disabled={sending || !draft.trim()}>
              Enviar
            </button>
          </form>
        </>
      )}
    </main>
  );
}
