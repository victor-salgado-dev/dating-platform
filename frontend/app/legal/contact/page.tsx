'use client';

import { FormEvent, useState } from 'react';

import { apiFetch, ApiError } from '@/lib/api';
import styles from '../legal.module.css';
import formStyles from '../../login/page.module.css';

export default function ContactPage() {
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [message, setMessage] = useState('');
  const [sent, setSent] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [sending, setSending] = useState(false);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setSending(true);
    setError(null);

    try {
      await apiFetch<void>('/contact', {
        method: 'POST',
        body: JSON.stringify({ name, email, message }),
      });
      setSent(true);
    } catch (err) {
      if (err instanceof ApiError && err.status === 429) {
        setError('Demasiados mensajes seguidos. Inténtalo de nuevo en un rato.');
      } else {
        setError('No se pudo enviar el mensaje. Inténtalo de nuevo.');
      }
    } finally {
      setSending(false);
    }
  }

  return (
    <main className={styles.main}>
      <h1>Contacto</h1>
      <p>
        ¿Dudas, problemas o quieres reportar algo grave? Escríbenos aquí. Para reportar
        a una persona concreta, hazlo directamente desde su perfil (botón
        &quot;Reportar&quot;) — así el equipo de moderación tiene todo el contexto.
      </p>

      {sent ? (
        <p>Gracias, hemos recibido tu mensaje. Te responderemos por email.</p>
      ) : (
        <form onSubmit={handleSubmit} className={formStyles.form}>
          <label>
            Nombre
            <input type="text" value={name} onChange={(e) => setName(e.target.value)} required />
          </label>
          <label>
            Email
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
            />
          </label>
          <label>
            Mensaje
            <textarea
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              required
              rows={5}
              maxLength={3000}
            />
          </label>

          {error && <p className={formStyles.error}>{error}</p>}

          <button type="submit" disabled={sending}>
            {sending ? 'Enviando…' : 'Enviar'}
          </button>
        </form>
      )}
    </main>
  );
}
