'use client';

import React, { useState } from 'react';
import { useRouter } from 'next/navigation';

import { useI18n } from '@/lib/i18n/context';
import styles from '../discover/page.module.css'; // Reutilizamos estilos

export default function SearchPage() {
  const router = useRouter();
  const { dictionary } = useI18n();
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
      <h1 className={styles.title}>{dictionary.search.title}</h1>

      <form
        onSubmit={handleSubmit}
        style={{
          maxWidth: '500px',
          margin: '0 auto',
          display: 'flex',
          flexDirection: 'column',
          gap: '1.5rem',
          padding: '2rem',
          backgroundColor: 'var(--surface-color, #fff)',
          borderRadius: '8px',
          border: '1px solid var(--border-color, #eaeaea)',
        }}
      >
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
          <label htmlFor="gender" style={{ fontWeight: 'bold' }}>
            {dictionary.search.labelGender}
          </label>
          <select
            name="gender"
            id="gender"
            value={formData.gender}
            onChange={handleChange}
            style={{ padding: '0.75rem', borderRadius: '4px', border: '1px solid #ccc' }}
          >
            <option value="">{dictionary.search.any}</option>
            <option value="female">{dictionary.search.genderFemale}</option>
            <option value="male">{dictionary.search.genderMale}</option>
            <option value="non_binary">{dictionary.search.genderNonBinary}</option>
            <option value="other">{dictionary.search.genderOther}</option>
          </select>
        </div>

        <div style={{ display: 'flex', gap: '1rem' }}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', flex: 1 }}>
            <label htmlFor="min_age" style={{ fontWeight: 'bold' }}>
              {dictionary.search.labelAgeMin}
            </label>
            <input
              type="number"
              name="min_age"
              id="min_age"
              placeholder={dictionary.search.placeholderAgeMin}
              min="18"
              max="99"
              value={formData.min_age}
              onChange={handleChange}
              style={{ padding: '0.75rem', borderRadius: '4px', border: '1px solid #ccc' }}
            />
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', flex: 1 }}>
            <label htmlFor="max_age" style={{ fontWeight: 'bold' }}>
              {dictionary.search.labelAgeMax}
            </label>
            <input
              type="number"
              name="max_age"
              id="max_age"
              placeholder={dictionary.search.placeholderAgeMax}
              min="18"
              max="99"
              value={formData.max_age}
              onChange={handleChange}
              style={{ padding: '0.75rem', borderRadius: '4px', border: '1px solid #ccc' }}
            />
          </div>
        </div>

        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
          <label htmlFor="relationship_goal" style={{ fontWeight: 'bold' }}>
            {dictionary.search.labelRelationshipGoal}
          </label>
          <select
            name="relationship_goal"
            id="relationship_goal"
            value={formData.relationship_goal}
            onChange={handleChange}
            style={{ padding: '0.75rem', borderRadius: '4px', border: '1px solid #ccc' }}
          >
            <option value="">{dictionary.search.any}</option>
            <option value="casual">{dictionary.search.goalCasual}</option>
            <option value="long_term">{dictionary.search.goalLongTerm}</option>
            <option value="friendship">{dictionary.search.goalFriendship}</option>
            <option value="marriage">{dictionary.search.goalMarriage}</option>
            <option value="not_sure">{dictionary.search.goalNotSure}</option>
          </select>
        </div>

        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
          <label htmlFor="country" style={{ fontWeight: 'bold' }}>
            {dictionary.search.labelCountry}
          </label>
          <input
            type="text"
            name="country"
            id="country"
            placeholder={dictionary.search.placeholderCountry}
            maxLength={2}
            value={formData.country.toUpperCase()}
            onChange={handleChange}
            style={{
              padding: '0.75rem',
              borderRadius: '4px',
              border: '1px solid #ccc',
              textTransform: 'uppercase',
            }}
          />
        </div>

        <button
          type="submit"
          style={{
            marginTop: '1rem',
            padding: '1rem',
            backgroundColor: 'var(--primary, #e60023)',
            color: 'white',
            border: 'none',
            borderRadius: '4px',
            fontSize: '1rem',
            cursor: 'pointer',
            fontWeight: 'bold',
          }}
        >
          {dictionary.search.submit}
        </button>
      </form>
    </main>
  );
}
