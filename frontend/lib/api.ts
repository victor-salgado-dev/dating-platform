// Helper mínimo para llamar a la API desde componentes cliente.
//
// Usa NEXT_PUBLIC_API_URL si está definido. Si no, usa una ruta relativa
// /api/v1 para que el navegador haga la petición al mismo origen desde el
// que se sirve la aplicación. Así la cookie httpOnly de sesión viaja
// automáticamente sin gestión manual y no saltan los 401 por CORS.
const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? '/api/v1';

export class ApiError extends Error {
  status: number;
  code?: string;

  constructor(status: number, message: string, code?: string) {
    super(message);
    this.status = status;
    this.code = code;
  }
}

// Duración de la caché de respuestas GET en milisegundos.
// Con 60 segundos evitamos la mayoría de duplicados al navegar entre páginas,
// sin mantener datos obsoletos durante mucho tiempo.
const GET_CACHE_TTL = 60_000;

// Caché de respuestas GET exitosas. Solo se usa en el navegador.
// La limpiamos después de cada mutación para que las siguientes GETs
// reflejen los cambios reales.
const getCache = new Map<string, { data: unknown; expiresAt: number }>();

function clearGetCache() {
  getCache.clear();
}

// Deduplicación de peticiones GET idénticas que están en curso.
// Si dos componentes intentan hacer la misma llamada al mismo tiempo,
// solo se enviará una request al backend.
const inFlightRequests = new Map<string, Promise<unknown>>();

async function doApiFetch<T>(path: string, init?: RequestInit): Promise<T> {
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

// Lectura síncrona de la caché de GET: devuelve la respuesta si sigue vigente.
// Permite pintar al instante lo que ya se pidió (p. ej. un perfil precargado)
// sin pasar por un estado de "cargando".
export function peekApiCache<T>(path: string): T | undefined {
  if (typeof window === 'undefined') return undefined;
  const cached = getCache.get(`GET ${path}`);
  if (cached && cached.expiresAt > Date.now()) return cached.data as T;
  return undefined;
}

// Mutación que NO vacía la caché de GET. Solo para escrituras que no cambian
// nada de lo que el usuario está viendo (registrar una visita): con apiFetch
// vaciaría, por ejemplo, el perfil que Quick Match acaba de precargar.
export async function apiMutateQuiet<T>(path: string, init: RequestInit): Promise<T> {
  return doApiFetch<T>(path, init);
}

export async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const method = (init?.method ?? 'GET').toUpperCase();

  if (method === 'GET') {
    const key = `GET ${path}`;

    // Con `cache: 'no-store'` se salta la caché de 60 s (útil para datos que
    // caducan rápido, como "online now"). La deduplicación en vuelo se mantiene.
    const skipCache = init?.cache === 'no-store';

    // Solo usamos la caché en el cliente. Así evitamos compartir estado
    // entre peticiones del servidor durante SSR/SSG.
    if (typeof window !== 'undefined' && !skipCache) {
      const cached = getCache.get(key);
      if (cached && cached.expiresAt > Date.now()) {
        return cached.data as T;
      }
      if (cached) {
        getCache.delete(key);
      }
    }

    const existing = inFlightRequests.get(key);
    if (existing) return existing as Promise<T>;

    const promise = doApiFetch<T>(path, init)
      .then((data) => {
        if (typeof window !== 'undefined' && !skipCache) {
          getCache.set(key, { data, expiresAt: Date.now() + GET_CACHE_TTL });
        }
        return data;
      })
      .finally(() => {
        inFlightRequests.delete(key);
      });

    inFlightRequests.set(key, promise);
    return promise;
  }

  // Para mutaciones POST/PUT/PATCH/DELETE no usamos la caché de GET.
  // Tras una mutación exitosa, limpiamos la caché para que los datos
  // posteriores estén actualizados.
  const result = await doApiFetch<T>(path, init);
  if (typeof window !== 'undefined') {
    clearGetCache();
  }
  return result;
}

// --- Tipos que reflejan las respuestas del backend (Fases 5 y 6) ---------

// Relación entre quien mira y el perfil de una ficha. Los calcula el backend
// en la misma consulta que el listado (ya no hay que descargar las listas de
// likes y favoritos para pintar los corazones y estrellas).
export type ProfileFlags = {
  liked: boolean; // yo le di like
  favorited: boolean; // lo tengo en favoritos
  received_like: boolean; // me dio like
  received_favorite: boolean; // me tiene en favoritos
};

export type SearchResultItem = {
  profile_id: string;
  display_name: string;
  age: number;
  gender: string;
  country_code: string;
  region: string | null;
  relationship_goal?: string | null;
  relationship_goals: string[] | null;
  has_photo: boolean;
  photo_url: string | null;
  created_at: string;
} & ProfileFlags;

export type SearchResponse = {
  items: SearchResultItem[];
  page: number;
  page_size: number;
  total: number;
  total_pages: number;
};

// /search/new-members, /search/online-now y /search/popular devuelven ahora
// exactamente lo mismo que /search/profiles. Se conservan los alias para no
// romper imports antiguos.
export type NewMemberItem = SearchResultItem;
export type NewMembersResponse = SearchResponse;
export type OnlineNowResponse = SearchResponse;
export type PopularResponse = SearchResponse;

// --- Listados de interacciones (likes, matches, favoritos, visitas, actividad)

type ListedProfile = {
  profile_id: string;
  display_name: string;
  age: number;
  gender: string;
  country_code: string;
  region: string | null;
  has_photo: boolean;
  photo_url: string | null;
} & ProfileFlags;

type Paged<T> = { items: T[]; page: number; page_size: number; total: number; total_pages: number };

export type LikeItem = ListedProfile & {
  relationship_goal: string | null;
  photo_id: string | null;
  liked_at: string;
};
export type LikesResponse = Paged<LikeItem>;

export type MatchItem = ListedProfile & {
  photo_id: string | null;
  matched_at: string;
  conversation_id?: string | null;
};
export type MatchesResponse = Paged<MatchItem>;

export type FavoriteItem = ListedProfile & {
  relationship_goal: string | null;
  photo_id: string | null;
  favorited_at: string;
};
export type FavoritesResponse = Paged<FavoriteItem>;

export type VisitItem = ListedProfile & {
  relationship_goal: string | null;
  photo_id: string | null;
  visited_at: string;
};
export type VisitsResponse = Paged<VisitItem>;

export type ActivityItem = ListedProfile & {
  event_type: 'like_received' | 'match_created' | 'favorite_received';
  created_at: string;
};
export type ActivityResponse = Paged<ActivityItem>;

// GET /auth/me: la cuenta más lo que necesita el header (antes eran tres peticiones).
export type MeResponse = {
  id: string;
  email: string;
  email_verified: boolean;
  status: string;
  created_at: string;
  has_profile: boolean;
  profile_id: string | null;
  photo_url: string | null;
  profile_completion: number; // 0-100
};

// POST /likes/{id}
export type LikeResult = { matched: boolean };

// PublicProfile refleja profileResponse del handler de profiles.
//
// OJO con dos cambios de contrato respecto a versiones anteriores:
//   - relationship_goal (string única) -> relationship_goals (array):
//     ahora se puede buscar más de una cosa a la vez.
//   - languages e interests YA NO están aquí: se gestionan aparte, con
//     sus propios endpoints (ver ProfileLanguage / ProfileInterest más
//     abajo), cada uno con nivel.
export type PublicProfile = {
  id: string;
  display_name: string;
  age: number;
  gender: string;
  country_code: string;
  region: string | null;
  relationship_goals: string[] | null;
  has_children: string | null;
  wants_children: string | null;
  bio: string | null;

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

  // --- Über mich / estilo de vida ---
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

// --- Idiomas ------------------------------------------------------------

export type ProfileLanguage = {
  language_code: string;
  level: number | null;
  updated_at: string;
};

// --- Intereses (sustituye a los hobbies de la Fase 2) -----------------------

export type InterestDefinition = {
  key: string;
  category: string;
  label: string;
  has_level: boolean;
  sort_order: number;
};

export type ProfileInterest = {
  interest_key: string;
  level: number | null;
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

// --- Respuesta consolidada de perfil completo -----------------------------
// Equivale a la respuesta de GET /profiles/{profileID}/full.

export type FullProfileEnvelope = {
  profile: PublicProfile;
  photos: ProfilePhoto[];
  languages: ProfileLanguage[];
  interests: ProfileInterest[];
  interest_catalog: InterestDefinition[];
  personality: PersonalityResponse;
  partner_preferences: PartnerPreferences;
  favorited: boolean;
  liked: boolean;
  matched: boolean;
  blocked: boolean;
};

// --- Mensajería ------------------------------------------------------------

export type MessageItem = {
  id: string;
  conversation_id: string;
  sender_id: string;
  body: string;
  created_at: string;
  delivered_at?: string | null;
  read_at?: string | null;
  is_mine: boolean;
};

export type MessagesResponse = Paged<MessageItem>;

// --- Reportes ----------------------------------------------------------------

export const REPORT_REASONS = [
  { value: 'spam', label: 'Spam' },
  { value: 'fake_profile', label: 'Perfil falso' },
  { value: 'harassment', label: 'Acoso' },
  { value: 'inappropriate_content', label: 'Contenido inapropiado' },
  { value: 'underage', label: 'Menor de edad' },
  { value: 'other', label: 'Otro' },
] as const;
