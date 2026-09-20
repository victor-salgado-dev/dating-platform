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

// PublicProfile refleja profileResponse del handler de profiles.
//
// OJO: has_children/wants_children son string (no boolean) desde la
// migración 000012 del backend ('yes' | 'no' | 'prefer_not_to_say' /
// 'not_sure'), no un booleano de tres estados. Si tenías código viejo
// tratándolos como boolean | null, hay que actualizarlo también.
export type PublicProfile = {
  id: string;
  display_name: string;
  age: number;
  gender: string;
  country_code: string;
  region: string | null;
  languages: string[] | null;
  relationship_goal: string | null;
  has_children: string | null;
  wants_children: string | null;
  bio: string | null;
  interests: string[] | null;

  // --- Físico y apariencia ---
  height: number | null;
  weight: number | null;
  body_type: string | null;
  ethnicity: string | null;
  appearance_rating: string | null;
  hair_color: string | null;
  eye_color: string | null;
  body_art: string[] | null;

  // --- Estilo de vida y familia ---
  smoking_habit: string | null;
  drinking_habit: string | null;
  relocation_willingness: string[] | null;
  marital_status: string | null;
  children_count: number | null;
  youngest_child_age: number | null;
  oldest_child_age: number | null;
  occupation: string | null;
  employment_status: string | null;
  income_level: string | null;
  living_situation: string | null;

  // --- Fondo, cultura y valores ---
  nationality: string | null;
  education_level: string | null;
  english_ability: string | null;
  religion: string | null;
  religious_values: string | null;
  star_sign: string | null;

  // --- Über mich / estilo de vida (Fase 1, migración 000013) ---
  future_vision: string[] | null;
  sports: string[] | null;
  likes_pets: string | null;
  pets_owned: string[] | null;
  favorite_season: string | null;
  ideal_vacation_style: string[] | null;
  vacation_activities: string[] | null;
  profile_quote: string | null;
  dream_wish: string | null;

  created_at: string;
  updated_at: string;
};

export type ProfilePhoto = {
  id: string;
  url: string;
  position: number;
  created_at: string;
};

// --- Hobbies (Fase 2) ------------------------------------------------------

export type HobbyDefinition = {
  key: string;
  category: string;
  label: string;
  sort_order: number;
};

export type ProfileHobby = {
  hobby_key: string;
  liked: boolean;
  intensity: number | null;
  updated_at: string;
};

// --- Personalidad (Fase 2) --------------------------------------------------

export type PersonalityStatement = {
  key: string;
  trait_key: string;
  label: string;
  sort_order: number;
};

export type PersonalityAnswer = {
  statement_key: string;
  score: number;
  updated_at: string;
};

export type PersonalityTraitScore = {
  trait_key: string;
  average_score: number;
  answered_count: number;
};

export type PersonalityResponse = {
  answers: PersonalityAnswer[];
  trait_scores: PersonalityTraitScore[];
};

// --- Preferencias de pareja (Fase 2) -----------------------------------------

export type PartnerPreferences = {
  age_min: number | null;
  age_max: number | null;
  height_min: number | null;
  height_max: number | null;
  desired_traits: string[] | null;
  partner_may_have_children: string | null;
  partner_religion_preference: string | null;
  about_partner_text: string | null;
  first_meeting_preference: string | null;
  desired_living_place: string[] | null;
  importance_shared_thoughts: number | null;
  importance_shared_hobbies: number | null;
  importance_intimacy: number | null;
  importance_romantic_love: number | null;
  importance_financial_security: number | null;
  importance_fun: number | null;
  importance_shared_friends: number | null;
  importance_shared_humor: number | null;
  importance_personal_space: number | null;
  importance_independence: number | null;
  updated_at?: string;
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
