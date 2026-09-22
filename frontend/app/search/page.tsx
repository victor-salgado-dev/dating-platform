'use client';

import React, { useState } from 'react';
import { useRouter } from 'next/navigation';
import styles from '../discover/page.module.css'; // Reutilizamos estilos

export default function SearchPage() {
  const router = useRouter();
  const [formData, setFormData] = useState({
    gender: '',
    min_age: '',
    max_age: '',
    relationship_goal: '',
    country: '',
  });

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    setFormData({ ...formData, [e.target.name]: e.target.value });
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const params = new URLSearchParams();

    // Solo añadimos los filtros que el usuario haya rellenado
    Object.entries(formData).forEach(([key, value]) => {
      if (value) params.append(key, value.trim());
    });

    // Redirigimos a descubrir con los filtros aplicados
    router.push(`/discover?${params.toString()}`);
  };

  return (
    <main className={styles.main}>
      <h1 className={styles.title}>Búsqueda Avanzada</h1>
      
      <form onSubmit={handleSubmit} style={{ maxWidth: '500px', margin: '0 auto', display: 'flex', flexDirection: 'column', gap: '1.5rem', padding: '2rem', backgroundColor: 'var(--surface-color, #fff)', borderRadius: '8px', border: '1px solid var(--border-color, #eaeaea)' }}>
        
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
          <label htmlFor="gender" style={{ fontWeight: 'bold' }}>Género que busco</label>
          <select name="gender" id="gender" value={formData.gender} onChange={handleChange} style={{ padding: '0.75rem', borderRadius: '4px', border: '1px solid #ccc' }}>
            <option value="">Cualquiera</option>
            <option value="female">Mujer</option>
            <option value="male">Hombre</option>
            <option value="non_binary">No binario</option>
            <option value="other">Otro</option>
          </select>
        </div>

        <div style={{ display: 'flex', gap: '1rem' }}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', flex: 1 }}>
            <label htmlFor="min_age" style={{ fontWeight: 'bold' }}>Edad mínima</label>
            <input type="number" name="min_age" id="min_age" placeholder="Ej: 18" min="18" max="99" value={formData.min_age} onChange={handleChange} style={{ padding: '0.75rem', borderRadius: '4px', border: '1px solid #ccc' }} />
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', flex: 1 }}>
            <label htmlFor="max_age" style={{ fontWeight: 'bold' }}>Edad máxima</label>
            <input type="number" name="max_age" id="max_age" placeholder="Ej: 50" min="18" max="99" value={formData.max_age} onChange={handleChange} style={{ padding: '0.75rem', borderRadius: '4px', border: '1px solid #ccc' }} />
          </div>
        </div>

        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
          <label htmlFor="relationship_goal" style={{ fontWeight: 'bold' }}>Tipo de relación</label>
          <select name="relationship_goal" id="relationship_goal" value={formData.relationship_goal} onChange={handleChange} style={{ padding: '0.75rem', borderRadius: '4px', border: '1px solid #ccc' }}>
            <option value="">Cualquiera</option>
            <option value="casual">Casual</option>
            <option value="long_term">Largo plazo</option>
            <option value="friendship">Amistad</option>
            <option value="marriage">Matrimonio</option>
            <option value="not_sure">No lo sé aún</option>
          </select>
        </div>

        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
          <label htmlFor="country" style={{ fontWeight: 'bold' }}>País (Código de 2 letras)</label>
          <input type="text" name="country" id="country" placeholder="Ej: ES, MX, AR..." maxLength={2} value={formData.country.toUpperCase()} onChange={handleChange} style={{ padding: '0.75rem', borderRadius: '4px', border: '1px solid #ccc', textTransform: 'uppercase' }} />
        </div>

        <button type="submit" style={{ marginTop: '1rem', padding: '1rem', backgroundColor: 'var(--primary, #e60023)', color: 'white', border: 'none', borderRadius: '4px', fontSize: '1rem', cursor: 'pointer', fontWeight: 'bold' }}>
          Buscar perfiles
        </button>
      </form>
    </main>
  );
}
