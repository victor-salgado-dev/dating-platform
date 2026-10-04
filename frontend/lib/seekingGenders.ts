// "Qué busco": el formulario ofrece tres opciones, pero el backend guarda los
// valores de Gender que ya existen (female, male, non_binary, other).
// "trans" envía non_binary y other juntos para que nadie quede invisible.
// Una única definición para registro y edición de perfil.

export type SeekingOptionId = 'male' | 'female' | 'trans';

export const SEEKING_OPTIONS: ReadonlyArray<{ id: SeekingOptionId; values: readonly string[] }> = [
  { id: 'male', values: ['male'] },
  { id: 'female', values: ['female'] },
  { id: 'trans', values: ['non_binary', 'other'] },
];

// Opciones marcadas -> valores para el backend (sin repetir).
export function seekingValuesFromOptions(selected: SeekingOptionId[]): string[] {
  const out: string[] = [];
  for (const opt of SEEKING_OPTIONS) {
    if (selected.includes(opt.id)) out.push(...opt.values);
  }
  return out;
}

// Valores del backend -> opciones marcadas. Una opción cuenta como marcada si
// tiene marcado al menos uno de sus valores.
export function seekingOptionsFromValues(values: string[] | null | undefined): SeekingOptionId[] {
  const set = new Set(values ?? []);
  return SEEKING_OPTIONS.filter((o) => o.values.some((v) => set.has(v))).map((o) => o.id);
}
