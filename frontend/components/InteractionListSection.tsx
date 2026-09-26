'use client';

// Componente genérico que reemplaza la lógica casi idéntica que tenían
// LikesSection, VisitsSection y FavoritesSection (pestañas + orden +
// grid + paginación). Cada sección solo aporta config: de dónde saca los
// datos de cada pestaña y los textos a mostrar. Así, añadir una cuarta
// lista de interacciones en el futuro es escribir un fetchPage, no
// duplicar otras 150 líneas.

import React, { useEffect, useState } from 'react';
import Link from 'next/link';
import { ProfileCard, ProfileItem } from './ProfileCard';
import { useProfileInteractions } from '@/lib/useProfileInteractions';
import { EmptyState, ErrorBanner } from './ListSectionState';
import styles from './ListSection.module.css';

export interface InteractionCardItem extends ProfileItem {
  interaction_at: string;
  conversation_id?: string | null;
}

export interface InteractionTabConfig<T extends string> {
  key: T;
  label: string;
  emptyTitle: string;
  emptySubtitle: string;
  fetchPage: (page: number) => Promise<{
    items: InteractionCardItem[];
    page: number;
    total_pages: number;
    total: number;
  }>;
}

export interface InteractionListSectionProps<T extends string> {
  tabs: InteractionTabConfig<T>[];
  defaultTab: T;
  loadingLabel: string;
  unauthorizedError: string;
  genericError: string;
  isUnauthorized: (err: unknown) => boolean;
  paginationInfo: (page: number, totalPages: number, total: number) => string;
  sortLabel: string;
  sortNewestFirst: string;
  sortOldestFirst: string;
  paginationPrev: string;
  paginationNext: string;
  // Solo Likes necesita el link a "abrir conversación" tras un match.
  openConversationLabel?: string;
}

export function InteractionListSection<T extends string>({
  tabs,
  defaultTab,
  loadingLabel,
  unauthorizedError,
  genericError,
  isUnauthorized,
  paginationInfo,
  sortLabel,
  sortNewestFirst,
  sortOldestFirst,
  paginationPrev,
  paginationNext,
  openConversationLabel,
}: InteractionListSectionProps<T>) {
  const [tab, setTab] = useState<T>(defaultTab);
  const [items, setItems] = useState<InteractionCardItem[]>([]);
  const [page, setPage] = useState(1);
  const [meta, setMeta] = useState({ page: 1, total_pages: 1, total: 0 });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [sortDesc, setSortDesc] = useState(true);

  const { likedIds, favoritedIds, receivedLikeIds, receivedFavIds, toggleLike, toggleFavorite } =
    useProfileInteractions();

  const activeConfig = tabs.find((t) => t.key === tab) ?? tabs[0];

  useEffect(() => {
    let active = true;
    setLoading(true);
    setError(null);

    activeConfig
      .fetchPage(page)
      .then((data) => {
        if (!active) return;
        setItems(data.items);
        setMeta({ page: data.page, total_pages: data.total_pages, total: data.total });
      })
      .catch((err: unknown) => {
        if (!active) return;
        setError(isUnauthorized(err) ? unauthorizedError : genericError);
      })
      .finally(() => {
        if (active) setLoading(false);
      });

    return () => {
      active = false;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [tab, page]);

  function changeTab(nextTab: T) {
    setTab(nextTab);
    setPage(1);
  }

  // Igual que antes: solo ordena lo ya cargado en esta página, no hay
  // sort_by/order en el backend todavía.
  const sortedItems = [...items].sort((a, b) => {
    const diff = new Date(b.interaction_at).getTime() - new Date(a.interaction_at).getTime();
    return sortDesc ? diff : -diff;
  });

  const totalPages = meta.total_pages;

  return (
    <div>
      <div className={styles.tabsRow}>
        <div className={styles.tabs} role="tablist">
          {tabs.map((t) => (
            <button
              key={t.key}
              type="button"
              className={tab === t.key ? styles.activeTab : styles.tab}
              onClick={() => changeTab(t.key)}
            >
              {t.label}
            </button>
          ))}
        </div>

        {items.length > 1 && (
          <div className={styles.sortRow}>
            <span>{sortLabel}</span>
            <button type="button" className={styles.sortButton} onClick={() => setSortDesc((v) => !v)}>
              {sortDesc ? sortNewestFirst : sortOldestFirst}
            </button>
          </div>
        )}
      </div>

      {loading && <p style={{ textAlign: 'center', padding: '2rem' }}>{loadingLabel}</p>}
      {error && <ErrorBanner message={error} className={styles.errorBanner} />}

      {!loading && !error && (
        <>
          {sortedItems.length === 0 ? (
            <EmptyState
              title={activeConfig.emptyTitle}
              subtitle={activeConfig.emptySubtitle}
              className={styles.emptyState}
              titleClassName={styles.emptyStateTitle}
              subtitleClassName={styles.emptyStateSubtitle}
            />
          ) : (
            <div className={styles.grid}>
              {sortedItems.map((item) => {
                const card = (
                  <ProfileCard
                    profile={item}
                    isPremium={false}
                    initialLiked={likedIds.has(item.profile_id)}
                    initialFavorited={favoritedIds.has(item.profile_id)}
                    receivedLike={receivedLikeIds.has(item.profile_id)}
                    receivedFavorite={receivedFavIds.has(item.profile_id)}
                    onToggleLike={toggleLike}
                    onToggleFavorite={toggleFavorite}
                  />
                );

                // Solo Likes muestra el link de conversación bajo la
                // tarjeta; el resto renderiza la tarjeta suelta, igual
                // que antes.
                if (openConversationLabel) {
                  return (
                    <div key={item.profile_id} className={styles.cardWrap}>
                      {card}
                      {item.conversation_id && (
                        <Link href={`/messages/${item.conversation_id}`} className={styles.openConversation}>
                          {openConversationLabel}
                        </Link>
                      )}
                    </div>
                  );
                }

                return <React.Fragment key={item.profile_id}>{card}</React.Fragment>;
              })}
            </div>
          )}

          {totalPages > 1 && (
            <div className={styles.pagination}>
              <button type="button" className={styles.pageButton} disabled={page <= 1} onClick={() => setPage((p) => Math.max(1, p - 1))}>
                {paginationPrev}
              </button>
              <span className={styles.pageInfo}>{paginationInfo(meta.page, totalPages, meta.total)}</span>
              <button type="button" className={styles.pageButton} disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)}>
                {paginationNext}
              </button>
            </div>
          )}
        </>
      )}
    </div>
  );
}
