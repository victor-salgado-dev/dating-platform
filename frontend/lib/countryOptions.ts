// Lista de códigos ISO 3166-1 alpha-2, para poblar selects de país /
// nacionalidad sin que el usuario tenga que escribir nada. El backend
// sigue validando country_code como ISO 3166-1 alpha-2 de siempre; esto
// es solo la lista de opciones del <select>.
//
// Las etiquetas legibles NO se guardan aquí ni en los diccionarios de
// i18n: se resuelven en tiempo de ejecución con Intl.DisplayNames, que
// el navegador ya trae y sabe traducir ~249 países a es/en (y a
// cualquier locale que se añada después) sin mantener esa lista a mano.

export const COUNTRY_CODES = [
  'AD', 'AE', 'AF', 'AG', 'AI', 'AL', 'AM', 'AO', 'AQ', 'AR', 'AS', 'AT', 'AU', 'AW', 'AX', 'AZ',
  'BA', 'BB', 'BD', 'BE', 'BF', 'BG', 'BH', 'BI', 'BJ', 'BL', 'BM', 'BN', 'BO', 'BQ', 'BR', 'BS',
  'BT', 'BV', 'BW', 'BY', 'BZ',
  'CA', 'CC', 'CD', 'CF', 'CG', 'CH', 'CI', 'CK', 'CL', 'CM', 'CN', 'CO', 'CR', 'CU', 'CV', 'CW',
  'CX', 'CY', 'CZ',
  'DE', 'DJ', 'DK', 'DM', 'DO', 'DZ',
  'EC', 'EE', 'EG', 'EH', 'ER', 'ES', 'ET',
  'FI', 'FJ', 'FK', 'FM', 'FO', 'FR',
  'GA', 'GB', 'GD', 'GE', 'GF', 'GG', 'GH', 'GI', 'GL', 'GM', 'GN', 'GP', 'GQ', 'GR', 'GS', 'GT',
  'GU', 'GW', 'GY',
  'HK', 'HM', 'HN', 'HR', 'HT', 'HU',
  'ID', 'IE', 'IL', 'IM', 'IN', 'IO', 'IQ', 'IR', 'IS', 'IT',
  'JE', 'JM', 'JO', 'JP',
  'KE', 'KG', 'KH', 'KI', 'KM', 'KN', 'KP', 'KR', 'KW', 'KY', 'KZ',
  'LA', 'LB', 'LC', 'LI', 'LK', 'LR', 'LS', 'LT', 'LU', 'LV', 'LY',
  'MA', 'MC', 'MD', 'ME', 'MF', 'MG', 'MH', 'MK', 'ML', 'MM', 'MN', 'MO', 'MP', 'MQ', 'MR', 'MS',
  'MT', 'MU', 'MV', 'MW', 'MX', 'MY', 'MZ',
  'NA', 'NC', 'NE', 'NF', 'NG', 'NI', 'NL', 'NO', 'NP', 'NR', 'NU', 'NZ',
  'OM',
  'PA', 'PE', 'PF', 'PG', 'PH', 'PK', 'PL', 'PM', 'PN', 'PR', 'PS', 'PT', 'PW', 'PY',
  'QA',
  'RE', 'RO', 'RS', 'RU', 'RW',
  'SA', 'SB', 'SC', 'SD', 'SE', 'SG', 'SH', 'SI', 'SJ', 'SK', 'SL', 'SM', 'SN', 'SO', 'SR', 'SS',
  'ST', 'SV', 'SX', 'SY', 'SZ',
  'TC', 'TD', 'TF', 'TG', 'TH', 'TJ', 'TK', 'TL', 'TM', 'TN', 'TO', 'TR', 'TT', 'TV', 'TW', 'TZ',
  'UA', 'UG', 'UM', 'US', 'UY', 'UZ',
  'VA', 'VC', 'VE', 'VG', 'VI', 'VN', 'VU',
  'WF', 'WS',
  'YE', 'YT',
  'ZA', 'ZM', 'ZW',
] as const;

let displayNamesCache: { locale: string; instance: Intl.DisplayNames } | null = null;

function getDisplayNames(locale: string): Intl.DisplayNames | null {
  if (typeof Intl === 'undefined' || typeof Intl.DisplayNames === 'undefined') return null;
  if (displayNamesCache?.locale === locale) return displayNamesCache.instance;
  try {
    const instance = new Intl.DisplayNames([locale], { type: 'region' });
    displayNamesCache = { locale, instance };
    return instance;
  } catch {
    return null;
  }
}

/** Nombre legible de un código ISO 3166-1 alpha-2 en el idioma dado. Si el
 * navegador no soporta Intl.DisplayNames (muy raro hoy en día), cae de
 * vuelta al propio código para no dejar el select vacío. */
export function getCountryLabel(code: string, locale: string): string {
  const dn = getDisplayNames(locale);
  return dn?.of(code) ?? code;
}

/** Códigos ordenados alfabéticamente por su nombre legible en el idioma
 * dado, para que el <select> no se quede en orden ISO (poco natural para
 * quien lo usa). */
export function sortedCountryOptions(locale: string): { code: string; label: string }[] {
  return COUNTRY_CODES.map((code) => ({ code, label: getCountryLabel(code, locale) })).sort((a, b) =>
    a.label.localeCompare(b.label, locale),
  );
}
