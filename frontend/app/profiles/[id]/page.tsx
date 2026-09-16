'use client';

import { useEffect, useState, FormEvent } from 'react';
import Link from 'next/link';
import { useParams, useRouter } from 'next/navigation';

import { apiFetch, ApiError, PublicProfile, ProfilePhoto, MessageItem, REPORT_REASONS } from '@/lib/api';
import styles from './page.module.css';

export default function ProfilePage() {
  const params = useParams<{ id: string }>();
  const router = useRouter();
  const [profile, setProfile] = useState<PublicProfile | null>(null);
  const [photos, setPhotos] = useState<ProfilePhoto[]>([]);
  const [favorited, setFavorited] = useState<boolean | null>(null);
  const [favoriteBusy, setFavoriteBusy] = useState(false);
  const [messageDraft, setMessageDraft] = useState('');
  const [sendingMessage, setSendingMessage] = useState(false);
  const [messageError, setMessageError] = useState<string | null>(null);
  const [blocked, setBlocked] = useState<boolean | null>(null);
  const [blockBusy, setBlockBusy] = useState(false);
  const [showReportForm, setShowReportForm] = useState(false);
  const [reportReason, setReportReason] = useState<string>(REPORT_REASONS[0].value);
  const [reportDescription, setReportDescription] = useState('');
  const [reportSending, setReportSending] = useState(false);
  const [reportSent, setReportSent] = useState(false);
  const [reportError, setReportError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!params?.id) return;

    setLoading(true);
    setError(null);

    apiFetch<PublicProfile>(`/profiles/${params.id}`)
      .then(async (p) => {
        setProfile(p);
        try {
          setPhotos(await apiFetch<ProfilePhoto[]>(`/profiles/${params.id}/photos`));
        } catch {
          // Si fallan las fotos no bloqueamos la vista del resto del perfil.
          setPhotos([]);
        }
        try {
          const status = await apiFetch<{ favorited: boolean }>(`/favorites/${params.id}`);
          setFavorited(status.favorited);
        } catch {
          // Si falla el estado de favorito, simplemente no mostramos el botón.
          setFavorited(null);
        }
        try {
          const status = await apiFetch<{ blocked: boolean }>(`/blocks/${params.id}`);
          setBlocked(status.blocked);
        } catch {
          setBlocked(null);
        }
      })
      .catch((err: unknown) => {
        if (err instanceof ApiError && err.status === 404) {
          setError('Este perfil no existe o ya no está disponible.');
        } else if (err instanceof ApiError && err.status === 401) {
          setError('Inicia sesión para ver este perfil.');
        } else {
          setError('No se pudo cargar el perfil.');
        }
      })
      .finally(() => setLoading(false));
  }, [params?.id]);

  async function toggleFavorite() {
    if (!params?.id || favorited === null || favoriteBusy) return;

    setFavoriteBusy(true);
    const next = !favorited;
    setFavorited(next); // optimista: el backend es idempotente, así que revertir en error es seguro

    try {
      await apiFetch<void>(`/favorites/${params.id}`, { method: next ? 'POST' : 'DELETE' });
    } catch {
      setFavorited(!next); // revertir si la llamada falla
    } finally {
      setFavoriteBusy(false);
    }
  }

  async function handleSendMessage(e: FormEvent) {
    e.preventDefault();
    if (!params?.id || !messageDraft.trim() || sendingMessage) return;

    setSendingMessage(true);
    setMessageError(null);
    try {
      const sent = await apiFetch<MessageItem>(`/messages/to/${params.id}`, {
        method: 'POST',
        body: JSON.stringify({ body: messageDraft }),
      });
      router.push(`/messages/${sent.conversation_id}`);
    } catch (err) {
      if (err instanceof ApiError && err.status === 400 && err.code === 'cannot_message_self') {
        setMessageError('No puedes enviarte un mensaje a ti mismo.');
      } else {
        setMessageError('No se pudo enviar el mensaje.');
      }
    } finally {
      setSendingMessage(false);
    }
  }

  async function toggleBlock() {
    if (!params?.id || blocked === null || blockBusy) return;

    setBlockBusy(true);
    const next = !blocked;

    try {
      await apiFetch<void>(`/blocks/${params.id}`, { method: next ? 'POST' : 'DELETE' });
      if (next) {
        // Una vez bloqueado, este perfil deja de ser visible para ti en
        // el resto de la app (regla de visibilidad de la Fase 9):
        // no tiene sentido quedarse en esta página.
        router.push('/discover');
        return;
      }
      setBlocked(false);
    } catch {
      // No revertimos optimistamente aquí: preferimos que el estado se
      // quede como estaba si la llamada falla.
    } finally {
      setBlockBusy(false);
    }
  }

  async function handleSendReport(e: FormEvent) {
    e.preventDefault();
    if (!params?.id || reportSending) return;

    setReportSending(true);
    setReportError(null);
    try {
      await apiFetch<void>(`/reports/${params.id}`, {
        method: 'POST',
        body: JSON.stringify({ reason: reportReason, description: reportDescription }),
      });
      setReportSent(true);
      setShowReportForm(false);
    } catch (err) {
      if (err instanceof ApiError && err.status === 400 && err.code === 'cannot_report_self') {
        setReportError('No puedes reportarte a ti mismo.');
      } else {
        setReportError('No se pudo enviar el reporte.');
      }
    } finally {
      setReportSending(false);
    }
  }

  return (
    <main className={styles.main}>
      <Link href="/discover" className={styles.back}>
        &larr; Volver a resultados
      </Link>

      {loading && <p>Cargando…</p>}
      {error && <p className={styles.error}>{error}</p>}

      {profile && (
        <article>
          <h1>
            {profile.display_name}, {profile.age}
          </h1>

          {favorited !== null && (
            <button
              type="button"
              onClick={toggleFavorite}
              disabled={favoriteBusy}
              className={favorited ? styles.favoriteActive : styles.favoriteButton}
            >
              {favorited ? '★ En favoritos' : '☆ Añadir a favoritos'}
            </button>
          )}

          <form onSubmit={handleSendMessage} className={styles.messageForm}>
            <input
              type="text"
              value={messageDraft}
              onChange={(e) => setMessageDraft(e.target.value)}
              placeholder={`Envía un mensaje a ${profile.display_name}…`}
              maxLength={2000}
              disabled={sendingMessage}
            />
            <button type="submit" disabled={sendingMessage || !messageDraft.trim()}>
              Enviar
            </button>
          </form>
          {messageError && <p className={styles.error}>{messageError}</p>}

          <p className={styles.location}>
            {[profile.region, profile.country_code].filter(Boolean).join(', ')}
          </p>

          <dl className={styles.details}>
            <dt>Género</dt>
            <dd>{profile.gender}</dd>
            <dt>¿Tiene hijos?</dt>
            <dd>{profile.has_children === null ? 'No indicado' : profile.has_children ? 'Sí' : 'No'}</dd>
            <dt>¿Quiere tener hijos?</dt>
            <dd>{profile.wants_children === null ? 'No indicado' : profile.wants_children ? 'Sí' : 'No'}</dd>
          </dl>

          {photos.length > 0 && (
            <div className={styles.photos}>
              {photos.map((photo) => (
                // Ruta relativa servida por el propio backend vía Caddy
                // (mismo origen): no hace falta anteponer NEXT_PUBLIC_API_URL.
                <img key={photo.id} src={photo.url} alt="" className={styles.photo} />
              ))}
            </div>
          )}

          {profile.bio && <p className={styles.bio}>{profile.bio}</p>}

          <dl className={styles.details}>
            {profile.relationship_goal && (
              <>
                <dt>Busca</dt>
                <dd>{profile.relationship_goal}</dd>
              </>
            )}
            {profile.languages && profile.languages.length > 0 && (
              <>
                <dt>Idiomas</dt>
                <dd>{profile.languages.join(', ')}</dd>
              </>
            )}
            {profile.interests && profile.interests.length > 0 && (
              <>
                <dt>Intereses</dt>
                <dd>{profile.interests.join(', ')}</dd>
              </>
            )}
          </dl>

          <div className={styles.safetyActions}>
            {blocked !== null && (
              <button type="button" onClick={toggleBlock} disabled={blockBusy}>
                {blocked ? 'Desbloquear' : 'Bloquear'}
              </button>
            )}
            {!reportSent && (
              <button type="button" onClick={() => setShowReportForm((v) => !v)}>
                Reportar
              </button>
            )}
            {reportSent && <span>Reporte enviado. Gracias.</span>}
          </div>

          {showReportForm && (
            <form onSubmit={handleSendReport} className={styles.reportForm}>
              <label>
                Motivo
                <select value={reportReason} onChange={(e) => setReportReason(e.target.value)}>
                  {REPORT_REASONS.map((r) => (
                    <option key={r.value} value={r.value}>
                      {r.label}
                    </option>
                  ))}
                </select>
              </label>
              <label>
                Detalles (opcional)
                <textarea
                  value={reportDescription}
                  onChange={(e) => setReportDescription(e.target.value)}
                  maxLength={2000}
                  rows={3}
                />
              </label>
              <button type="submit" disabled={reportSending}>
                Enviar reporte
              </button>
            </form>
          )}
          {reportError && <p className={styles.error}>{reportError}</p>}
        </article>
      )}
    </main>
  );
}
