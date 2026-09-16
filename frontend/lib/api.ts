// Helper mínimo para llamar a la API desde componentes cliente.
//
// Usa NEXT_PUBLIC_API_URL (el navegador llega a la API a través de Caddy,
// en el mismo origen que la propia página), así que la cookie de sesión
// httpOnly viaja automáticamente sin ninguna gestión manual aquí.
const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost/api/v1';

export class ApiError extends Error {
  status: number;
  code?: string;

  constructor(status: number, message: string, code?: string) {
    super(message);
    this.status = status;
    this.code = code;
  }
}

export async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const isFormData = typeof FormData !== 'undefined' && init?.body instanceof FormData;
  const res = await fetch(`${API_BASE}${path}`, {
    ...init,
    credentials: 'include',
    headers: {
      ...(isFormData ? {} : { 'Content-Type': 'application/json' }),
      ...(init?.headers ?? {}),
    },
  });

  if (!res.ok) {
    let message = `Error ${res.status}`;
    let code: string | undefined;
    try {
      const data = await res.json();
      message = data?.error?.message ?? message;
      code = data?.error?.code;
    } catch {
      // El cuerpo no era JSON: nos quedamos con el mensaje genérico.
    }
    throw new ApiError(res.status, message, code);
  }

  if (res.status === 204) {
    return undefined as T;
  }
  return (await res.json()) as T;
}

// --- Tipos que reflejan las respuestas del backend (Fases 5 y 6) ---------

export type SearchResultItem = {
  profile_id: string;
  display_name: string;
  age: number;
  gender: string;
  country_code: string;
  region: string | null;
  relationship_goal: string | null;
  has_photo: boolean;
  created_at: string;
};

export type SearchResponse = {
  items: SearchResultItem[];
  page: number;
  page_size: number;
  total: number;
  total_pages: number;
};

export type PublicProfile = {
  id: string;
  display_name: string;
  age: number;
  gender: string;
  country_code: string;
  region: string | null;
  languages: string[] | null;
  relationship_goal: string | null;
  has_children: boolean | null;
  wants_children: boolean | null;
  bio: string | null;
  interests: string[] | null;
  created_at: string;
  updated_at: string;
};

export type ProfilePhoto = {
  id: string;
  url: string;
  position: number;
  created_at: string;
};

export type FavoriteItem = {
  profile_id: string;
  display_name: string;
  age: number;
  gender: string;
  country_code: string;
  region: string | null;
  relationship_goal: string | null;
  has_photo: boolean;
  favorited_at: string;
};

export type FavoritesResponse = {
  items: FavoriteItem[];
  page: number;
  page_size: number;
  total: number;
  total_pages: number;
};

export type LikeItem = {
  profile_id: string;
  display_name: string;
  age: number;
  gender: string;
  country_code: string;
  region: string | null;
  relationship_goal: string | null;
  has_photo: boolean;
  liked_at: string;
};

export type LikesResponse = {
  items: LikeItem[];
  page: number;
  page_size: number;
  total: number;
  total_pages: number;
};

export type MatchItem = {
  profile_id: string;
  display_name: string;
  age: number;
  gender: string;
  country_code: string;
  region: string | null;
  has_photo: boolean;
  matched_at: string;
  conversation_id?: string;
};

export type MatchesResponse = {
  items: MatchItem[];
  page: number;
  page_size: number;
  total: number;
  total_pages: number;
};

export type ActivityItem = {
  event_type: 'like_received' | 'match_created' | 'favorite_received';
  profile_id: string;
  display_name: string;
  age: number;
  gender: string;
  country_code: string;
  region: string | null;
  has_photo: boolean;
  created_at: string;
};

export type ActivityResponse = {
  items: ActivityItem[];
  page: number;
  page_size: number;
  total: number;
  total_pages: number;
};

export type ConversationParticipant = {
  profile_id: string;
  display_name: string;
  age: number;
  gender: string;
  country_code: string;
  region: string | null;
  has_photo: boolean;
};

export type ConversationItem = {
  conversation_id: string;
  other_participant: ConversationParticipant;
  last_message: { body: string; created_at: string; is_mine: boolean };
  unread_count: number;
};

export type ConversationsResponse = {
  items: ConversationItem[];
  page: number;
  page_size: number;
  total: number;
  total_pages: number;
};

export type MessageItem = {
  id: string;
  conversation_id: string;
  body: string;
  created_at: string;
  is_mine: boolean;
  read_at: string | null;
};

export type MessagesResponse = {
  conversation_id: string;
  items: MessageItem[];
  page: number;
  page_size: number;
  total: number;
  total_pages: number;
};

export type BlockedItem = {
  profile_id: string;
  display_name: string;
  age: number;
  gender: string;
  country_code: string;
  region: string | null;
  blocked_at: string;
};

export type BlockedResponse = {
  items: BlockedItem[];
  page: number;
  page_size: number;
  total: number;
  total_pages: number;
};

export const REPORT_REASONS = [
  { value: 'spam', label: 'Spam' },
  { value: 'fake_profile', label: 'Perfil falso' },
  { value: 'harassment', label: 'Acoso' },
  { value: 'inappropriate_content', label: 'Contenido inapropiado' },
  { value: 'underage', label: 'Menor de edad' },
  { value: 'other', label: 'Otro' },
] as const;

// --- Administración (Fase 10) --------------------------------------------

export type AdminUser = {
  id: string;
  email: string;
  status: string;
  role: string;
  email_verified: boolean;
  created_at: string;
  deleted_at: string | null;
};

export type AdminUsersResponse = {
  items: AdminUser[];
  page: number;
  page_size: number;
  total: number;
  total_pages: number;
};

export type AdminReport = {
  id: string;
  reporter_id: string;
  reporter_name: string | null;
  reported_id: string;
  reported_name: string | null;
  reason: string;
  description: string | null;
  status: string;
  created_at: string;
};

export type AdminReportsResponse = {
  items: AdminReport[];
  page: number;
  page_size: number;
  total: number;
  total_pages: number;
};
