'use client';

import { apiFetch, ApiError, LikesResponse, MatchesResponse, LikeItem, MatchItem } from '@/lib/api';
import { useI18n } from '@/lib/i18n/context';
import { InteractionListSection, InteractionCardItem, InteractionTabConfig } from '@/components/InteractionListSection';

type Tab = 'received' | 'sent' | 'mutual';

function mapItem(item: LikeItem | MatchItem): InteractionCardItem {
  return {
    profile_id: item.profile_id,
    display_name: item.display_name,
    age: item.age,
    gender: item.gender,
    country_code: item.country_code,
    region: item.region,
    has_photo: item.has_photo,
    photo_url: item.photo_url,
    // liked_at para received/sent, matched_at para mutual (reusa /matches,
    // no existe /likes/mutual). Normalizado a un solo campo para ordenar.
    interaction_at: 'liked_at' in item ? item.liked_at : item.matched_at,
    conversation_id: 'conversation_id' in item ? item.conversation_id ?? null : null,
  };
}

export default function LikesSection() {
  const { dictionary } = useI18n();

  async function fetchTab(tab: Tab, page: number) {
    const path = tab === 'mutual' ? `/matches?page=${page}` : `/likes/${tab}?page=${page}`;
    const data = await apiFetch<LikesResponse | MatchesResponse>(path);
    return {
      items: (data.items as Array<LikeItem | MatchItem>).map(mapItem),
      page: data.page,
      total_pages: data.total_pages,
      total: data.total,
    };
  }

  const tabs: InteractionTabConfig<Tab>[] = [
    {
      key: 'received',
      label: dictionary.likes.tabReceived,
      emptyTitle: dictionary.likes.emptyReceived,
      emptySubtitle: dictionary.likes.emptySubtitle,
      fetchPage: (page) => fetchTab('received', page),
    },
    {
      key: 'sent',
      label: dictionary.likes.tabSent,
      emptyTitle: dictionary.likes.emptySent,
      emptySubtitle: dictionary.likes.emptySubtitle,
      fetchPage: (page) => fetchTab('sent', page),
    },
    {
      key: 'mutual',
      label: dictionary.likes.tabMutual,
      emptyTitle: dictionary.likes.emptyMutual,
      emptySubtitle: dictionary.likes.emptyMutualSubtitle,
      fetchPage: (page) => fetchTab('mutual', page),
    },
  ];

  return (
    <InteractionListSection
      tabs={tabs}
      defaultTab="received"
      loadingLabel={dictionary.likes.loading}
      unauthorizedError={dictionary.likes.errorUnauthorized}
      genericError={dictionary.likes.loadError}
      isUnauthorized={(err) => err instanceof ApiError && err.status === 401}
      paginationInfo={(page, totalPages, total) =>
        dictionary.likes.paginationInfo
          .replace('{page}', String(page))
          .replace('{totalPages}', String(totalPages))
          .replace('{total}', String(total))
      }
      sortLabel={dictionary.common.sortLabel}
      sortNewestFirst={dictionary.common.sortNewestFirst}
      sortOldestFirst={dictionary.common.sortOldestFirst}
      paginationPrev={dictionary.common.paginationPrev}
      paginationNext={dictionary.common.paginationNext}
      openConversationLabel={dictionary.likes.openConversation}
    />
  );
}
