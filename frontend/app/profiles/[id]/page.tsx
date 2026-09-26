'use client';

import { useEffect, useState, type FormEvent } from 'react';
import Link from 'next/link';
import { useParams, useRouter } from 'next/navigation';

import { apiFetch, ApiError, MessageItem, REPORT_REASONS } from '@/lib/api';
import { useI18n } from '@/lib/i18n/context';
import { useFullProfile } from '@/lib/useFullProfile';
import { FullProfileSections } from '@/components/FullProfileSections';
import styles from './page.module.css';

const REPORT_REASON_KEY: Record<
  string,
  'spam' | 'fakeProfile' | 'harassment' | 'inappropriateContent' | 'underage' | 'other'
> = {
  spam: 'spam',
  fake_profile: 'fakeProfile',
  harassment: 'harassment',
  inappropriate_content: 'inappropriateContent',
  underage: 'underage',
  other: 'other',
};

export default function ProfilePage() {
  const params = useParams<{ id: string }>();
  const router = useRouter();
  const { dictionary } = useI18n();

  const { profile, photos, languages, interestCatalog, theirInterests, personality, partnerPrefs, loading, notFound, unauthorized } =
    useFullProfile(params?.id);

  const [favorited, setFavorited] = useState<boolean | null>(null);
  const [favoriteBusy, setFavoriteBusy] = useState(false);
  const [liked, setLiked] = useState<boolean | null>(null);
  const [matched, setMatched] = useState(false);
  const [likeBusy, setLikeBusy] = useState(false);
  const [showMatchNotice, setShowMatchNotice] = useState(false);
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

  // Registrar visita + estado de like/favorito/bloqueo. Separado de
  // useFullProfile porque esto es interacción del visitante, no datos del
  // perfil visitado (Quick Match no necesita nada de esto).
  useEffect(() => {
    if (!params?.id) return;
    apiFetch(`/visits/${params.id}`, { method: 'POST' }).catch(() => {});
    apiFetch<{ favorited: boolean }>(`/favorites/${params.id}`).then((s) => setFavorited(s.favorited)).catch(() => setFavorited(null));
    apiFetch<{ liked: boolean; matched: boolean }>(`/likes/${params.id}`)
      .then((s) => { setLiked(s.liked); setMatched(s.matched); })
      .catch(() => setLiked(null));
    apiFetch<{ blocked: boolean }>(`/blocks/${params.id}`).then((s) => setBlocked(s.blocked)).catch(() => setBlocked(null));
  }, [params?.id]);

  async function toggleFavorite() {
    if (!params?.id || favorited === null || favoriteBusy) return;
    setFavoriteBusy(true);
    const next = !favorited;
    setFavorited(next);
    try {
      await apiFetch<void>(`/favorites/${params.id}`, { method: next ? 'POST' : 'DELETE' });
    } catch {
      setFavorited(!next);
    } finally {
      setFavoriteBusy(false);
    }
  }

  async function toggleLike() {
    if (!params?.id || liked === null || likeBusy) return;
    setLikeBusy(true);
    const next = !liked;
    try {
      await apiFetch<void>(`/likes/${params.id}`, { method: next ? 'POST' : 'DELETE' });
      setLiked(next);
      if (next) {
        const status = await apiFetch<{ matched: boolean }>(`/matches/${params.id}`);
        setMatched(status.matched);
        if (status.matched && !matched) setShowMatchNotice(true);
      } else {
        setMatched(false);
      }
    } catch {
    } finally {
      setLikeBusy(false);
    }
  }

  async function handleSendMessage(e: FormEvent) {
    e.preventDefault();
    if (!params?.id || !messageDraft.trim() || sendingMessage || !profile) return;
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
        setMessageError(dictionary.profilePublic.messageCannotSelf);
      } else {
        setMessageError(dictionary.profilePublic.messageError);
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
        router.push('/discover');
        return;
      }
      setBlocked(false);
    } catch {
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
        setReportError(dictionary.profilePublic.reportCannotSelf);
      } else {
        setReportError(dictionary.profilePublic.reportError);
      }
    } finally {
      setReportSending(false);
    }
  }

  const error = notFound
    ? dictionary.profilePublic.notFound
    : unauthorized
    ? dictionary.profilePublic.unauthorized
    : !loading && !profile
    ? dictionary.profilePublic.loadError
    : null;

  return (
    <main className={styles.main}>
      <Link href="/discover" className={styles.back}>
        {dictionary.common.backToResults}
      </Link>

      {loading && <p>{dictionary.common.loading}</p>}
      {error && <p className={styles.error}>{error}</p>}

      {profile && (
        <article>
          <h1>
            {profile.display_name}, {profile.age}
          </h1>

          <p className={styles.location}>
            {[profile.region, profile.country_code].filter(Boolean).join(', ')}
          </p>

          {favorited !== null && (
            <button type="button" onClick={toggleFavorite} disabled={favoriteBusy} className={favorited ? styles.favoriteActive : styles.favoriteButton}>
              {favorited ? dictionary.profilePublic.favoriteActive : dictionary.profilePublic.favoriteInactive}
            </button>
          )}

          {liked !== null && (
            <button type="button" onClick={toggleLike} disabled={likeBusy} className={liked ? styles.likeActive : styles.likeButton}>
              {liked ? dictionary.profilePublic.likeActive : dictionary.profilePublic.likeInactive}
            </button>
          )}

          {showMatchNotice && (
            <div className={styles.matchNotice} role="status">
              {dictionary.profilePublic.matchNotice.replace('{name}', profile.display_name)}
              <button type="button" onClick={() => setShowMatchNotice(false)}>
                {dictionary.common.close}
              </button>
            </div>
          )}

          <form onSubmit={handleSendMessage} className={styles.messageForm}>
            <input
              type="text"
              value={messageDraft}
              onChange={(e) => setMessageDraft(e.target.value)}
              placeholder={dictionary.profilePublic.messagePlaceholder.replace('{name}', profile.display_name)}
              maxLength={2000}
              disabled={sendingMessage}
            />
            <button type="submit" disabled={sendingMessage || !messageDraft.trim()}>
              {dictionary.common.send}
            </button>
          </form>
          {messageError && <p className={styles.error}>{messageError}</p>}

          {photos.length > 0 && (
            <div className={styles.photos}>
              {photos.map((photo) => (
                <img key={photo.id} src={photo.url} alt="" className={styles.photo} loading="lazy" />
              ))}
            </div>
          )}

          <FullProfileSections
            profile={profile}
            languages={languages}
            interestCatalog={interestCatalog}
            theirInterests={theirInterests}
            personality={personality}
            partnerPrefs={partnerPrefs}
          />

          <div className={styles.safetyActions}>
            {blocked !== null && (
              <button type="button" onClick={toggleBlock} disabled={blockBusy}>
                {blocked ? dictionary.common.unblock : dictionary.common.block}
              </button>
            )}
            {!reportSent && (
              <button type="button" onClick={() => setShowReportForm((v) => !v)}>
                {dictionary.profilePublic.report}
              </button>
            )}
            {reportSent && <span>{dictionary.profilePublic.reportSent}</span>}
          </div>

          {showReportForm && (
            <form onSubmit={handleSendReport} className={styles.reportForm}>
              <label>
                {dictionary.profilePublic.reportReason}
                <select value={reportReason} onChange={(e) => setReportReason(e.target.value)}>
                  {REPORT_REASONS.map((r) => (
                    <option key={r.value} value={r.value}>
                      {dictionary.reports.reasons[REPORT_REASON_KEY[r.value] ?? 'other']}
                    </option>
                  ))}
                </select>
              </label>
              <label>
                {dictionary.profilePublic.reportDetails}
                <textarea value={reportDescription} onChange={(e) => setReportDescription(e.target.value)} maxLength={2000} rows={3} />
              </label>
              <button type="submit" disabled={reportSending}>
                {dictionary.profilePublic.reportSubmit}
              </button>
            </form>
          )}
          {reportError && <p className={styles.error}>{reportError}</p>}
        </article>
      )}
    </main>
  );
}
