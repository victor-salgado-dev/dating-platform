// Autocompletado de ubicación usando Photon (komoot.io), construida
// sobre datos de OpenStreetMap. Gratuita, sin registro ni API key. No
// es la fuente de verdad de nada: solo ayuda a rellenar region/
// country_code con un topónimo real en vez de que el usuario los
// escriba a mano — el backend sigue validando country_code como
// ISO 3166-1 alpha-2 de siempre.
//
// Límites a tener en cuenta si esto crece: el servicio público de
// Photon no da garantías de SLA ni de rate limit documentado. Para
// producción con tráfico serio, lo normal sería mover esto a una
// instancia propia de Photon/Nominatim o a un proveedor de pago — pero
// para arrancar (y mientras la app sea pequeña) esto no cuesta nada y
// funciona bien.

export type PlaceSuggestion = {
  /** Texto completo para mostrar en la sugerencia, ej. "Darmstadt, Hesse, Alemania" */
  label: string;
  city?: string;
  state?: string;
  country?: string;
  /** ISO 3166-1 alpha-2, en mayúsculas (ej. "DE") */
  countryCode: string;
};

const PHOTON_URL = 'https://photon.komoot.io/api/';

export async function searchPlaces(query: string): Promise<PlaceSuggestion[]> {
  const trimmed = query.trim();
  if (trimmed.length < 2) return [];

  const url = `${PHOTON_URL}?q=${encodeURIComponent(trimmed)}&limit=6&lang=en`;

  let res: Response;
  try {
    res = await fetch(url);
  } catch {
    // Sin conexión, o el servicio no responde: no rompemos el formulario
    // por esto, simplemente no hay sugerencias.
    return [];
  }
  if (!res.ok) return [];

  const data = await res.json().catch(() => null);
  const features: unknown[] = Array.isArray(data?.features) ? data.features : [];

  const suggestions: PlaceSuggestion[] = [];
  for (const feature of features) {
    const props = (feature as { properties?: Record<string, unknown> })?.properties;
    const countryCode = props?.countrycode;
    if (typeof countryCode !== 'string' || countryCode.length !== 2) {
      // Sin código de país no podemos rellenar country_code; se descarta.
      continue;
    }

    const name = typeof props?.name === 'string' ? props.name : undefined;
    const city = typeof props?.city === 'string' ? props.city : name;
    const state = typeof props?.state === 'string' ? props.state : undefined;
    const country = typeof props?.country === 'string' ? props.country : undefined;

    const label = [city ?? name, state, country].filter(Boolean).join(', ');
    if (!label) continue;

    suggestions.push({ label, city, state, country, countryCode: countryCode.toUpperCase() });
  }

  return suggestions;
}
