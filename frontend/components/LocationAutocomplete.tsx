'use client';

import { useEffect, useRef, useState } from 'react';

import { searchPlaces, PlaceSuggestion } from '@/lib/geocoding';
import styles from './LocationAutocomplete.module.css';

type Props = {
  /** Texto inicial a mostrar en el input (ej. "Darmstadt, Hesse" si ya lo tenía guardado) */
  initialValue?: string;
  onSelect: (place: PlaceSuggestion) => void;
};

const DEBOUNCE_MS = 300;

export default function LocationAutocomplete({ initialValue, onSelect }: Props) {
  const [query, setQuery] = useState(initialValue ?? '');
  const [suggestions, setSuggestions] = useState<PlaceSuggestion[]>([]);
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const requestIDRef = useRef(0);

  useEffect(() => {
    if (debounceRef.current) clearTimeout(debounceRef.current);

    const trimmed = query.trim();
    if (trimmed.length < 2) {
      setSuggestions([]);
      setOpen(false);
      return;
    }

    debounceRef.current = setTimeout(async () => {
      const thisRequest = ++requestIDRef.current;
      setLoading(true);
      const results = await searchPlaces(trimmed);
      // Si mientras esperábamos la respuesta el usuario siguió tecleando,
      // esta respuesta ya está obsoleta: se descarta.
      if (thisRequest === requestIDRef.current) {
        setSuggestions(results);
        setOpen(results.length > 0);
        setLoading(false);
      }
    }, DEBOUNCE_MS);

    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current);
    };
  }, [query]);

  function handleSelect(place: PlaceSuggestion) {
    setQuery(place.label);
    setSuggestions([]);
    setOpen(false);
    onSelect(place);
  }

  return (
    <div className={styles.wrapper}>
      <input
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        onFocus={() => suggestions.length > 0 && setOpen(true)}
        onBlur={() => setTimeout(() => setOpen(false), 150)}
        placeholder="Empieza a escribir una ciudad..."
        autoComplete="off"
      />
      {loading && <span className={styles.loading}>Buscando…</span>}
      {open && suggestions.length > 0 && (
        <ul className={styles.suggestions}>
          {suggestions.map((s, i) => (
            <li key={`${s.label}-${i}`}>
              <button type="button" onMouseDown={(e) => e.preventDefault()} onClick={() => handleSelect(s)}>
                {s.label}
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
