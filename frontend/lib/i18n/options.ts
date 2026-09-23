import type { Dictionary } from './dictionaries/es';

// Resuelve la etiqueta traducida de un valor de `profileOptions.ts` a
// partir de la categoría (nombre del array en camelCase) y del valor
// técnico. Si no hay traducción, devuelve el valor crudo — es un fallback
// silencioso para no romper la UI si algún día el backend añade un valor
// sin que lo hayamos registrado en el diccionario.
export function tOption(
  dictionary: Dictionary,
  category: keyof Dictionary['options'],
  value: string | null | undefined,
): string | null {
  if (!value) return null;
  const bucket = dictionary.options[category] as Record<string, string> | undefined;
  return bucket?.[value] ?? value;
}

export function tOptionList(
  dictionary: Dictionary,
  category: keyof Dictionary['options'],
  values: string[] | null | undefined,
): string | null {
  if (!values || values.length === 0) return null;
  return values.map((v) => tOption(dictionary, category, v) ?? v).join(', ');
}
